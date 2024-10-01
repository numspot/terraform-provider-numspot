package vm

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VmResource{}
	_ resource.ResourceWithConfigure   = &VmResource{}
	_ resource.ResourceWithImportState = &VmResource{}
)

type VmResource struct {
	provider services.IProvider
}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

func (r *VmResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(services.IProvider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)
		return
	}

	r.provider = provider
}

func (r *VmResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VmResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vm"
}

func (r *VmResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VmResourceSchema(ctx)
}

func (r *VmResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data VmModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		VmFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		r.provider.GetNumspotClient().CreateVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", err.Error())
	}
	if response.Diagnostics.HasError() {
		return
	}

	vm := *res.JSON201
	createdId := *vm.Id

	// Create tags
	if len(data.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.GetNumspotClient(), r.provider.GetSpaceID(), &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read, err := utils2.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"running", "stopped"}, // In some cases, when there is insufficient capacity the VM is created with state = stopped
		r.provider.GetNumspotClient().ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", createdId, err))
		return
	}

	vmSchema, ok := read.(*numspot.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to create VM", "object conversion error")
		return
	}

	// In some cases, when there is insufficient capacity the VM is created with state = stopped
	if utils2.GetPtrValue(vmSchema.State) == "stopped" {
		response.Diagnostics.AddError("Issue while creating VM", fmt.Sprintf("VM was created in 'stopped' state. Reason : %s", utils2.GetPtrValue(vmSchema.StateReason)))
		return
	}

	tf, diagnostics := VmFromHttpToTf(ctx, vmSchema)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils2.ExecuteRequest(func() (*numspot.ReadVmsByIdResponse, error) {
		id := utils2.FromTfStringToStringPtr(data.Id)
		if id == nil {
			return nil, errors.New("Found invalid id")
		}
		return r.provider.GetNumspotClient().ReadVmsByIdWithResponse(ctx, r.provider.GetSpaceID(), *id)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diagnostics := VmFromHttpToTf(ctx, res.JSON200)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VmModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vmId := state.Id.ValueString()

	// Update tags
	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.GetNumspotClient(),
			r.provider.GetSpaceID(),
			vmId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	body := VmFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	bodyFromState := VmFromTfToUpdaterequest(ctx, &state, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	if isUpdateNeeded(body, bodyFromState) {
		// Stop VM before doing update
		diags := StopVm(ctx, r.provider, vmId)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		// Update VM
		updatedRes := utils2.ExecuteRequest(func() (*numspot.UpdateVmResponse, error) {
			return r.provider.GetNumspotClient().UpdateVmWithResponse(ctx, r.provider.GetSpaceID(), vmId, body)
		}, http.StatusOK, &response.Diagnostics)

		if updatedRes == nil || response.Diagnostics.HasError() {
			return
		}

		// Restart VM
		diags = StartVm(ctx, r.provider, vmId)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	// Retries read on VM until state is OK
	read, err := utils2.RetryReadUntilStateValid(
		ctx,
		vmId,
		r.provider.GetSpaceID(),
		[]string{"pending"},
		[]string{"running"},
		r.provider.GetNumspotClient().ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to update VM", fmt.Sprintf("Error waiting for VM to be created: %s", err))
		return
	}
	vmObject, ok := read.(*numspot.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to update VM", "object conversion error")
		return
	}

	tf, diags := VmFromHttpToTf(ctx, vmObject)

	if diags.HasError() {
		response.Diagnostics.Append(diags...)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func compareSimpleFieldPtr[R comparable](val1 *R, val2 *R) bool {
	return utils2.GetPtrValue(val1) == utils2.GetPtrValue(val2)
}

func compareSlicePtr[R comparable](val1 *[]R, val2 *[]R) bool {
	return slices.Equal(utils2.GetPtrValue(val1), utils2.GetPtrValue(val2))
}

func isUpdateNeeded(plan numspot.UpdateVmJSONRequestBody, state numspot.UpdateVmJSONRequestBody) bool {
	return !(compareSimpleFieldPtr(plan.BsuOptimized, state.BsuOptimized) &&
		compareSimpleFieldPtr(plan.DeletionProtection, state.DeletionProtection) &&
		compareSimpleFieldPtr(plan.KeypairName, state.KeypairName) &&
		compareSimpleFieldPtr(plan.NestedVirtualization, state.NestedVirtualization) &&
		(utils2.GetPtrValue(plan.Performance) == "" || compareSimpleFieldPtr(plan.Performance, state.Performance)) && // if performance is not provided by user,
		(len(utils2.GetPtrValue(plan.BlockDeviceMappings)) == 0) &&
		compareSimpleFieldPtr(plan.UserData, state.UserData) &&
		compareSimpleFieldPtr(plan.VmInitiatedShutdownBehavior, state.VmInitiatedShutdownBehavior) &&
		compareSimpleFieldPtr(plan.Type, state.Type) &&
		(len(utils2.GetPtrValue(plan.BlockDeviceMappings)) == 0 || (compareSlicePtr(plan.BlockDeviceMappings, state.BlockDeviceMappings))) &&
		compareSimpleFieldPtr(plan.IsSourceDestChecked, state.IsSourceDestChecked))
}

func (r *VmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := utils2.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VM", err.Error())
		return
	}
}
