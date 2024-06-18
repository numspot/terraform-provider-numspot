package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VmResource{}
	_ resource.ResourceWithConfigure   = &VmResource{}
	_ resource.ResourceWithImportState = &VmResource{}
)

type VmResource struct {
	provider Provider
}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

func (r *VmResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
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
	response.Schema = resource_vm.VmResourceSchema(ctx)
}

func (r *VmResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		VmFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		r.provider.ApiClient.CreateVmsWithResponse)
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
		tags.CreateTagsFromTf(ctx, r.provider.ApiClient, r.provider.SpaceID, &response.Diagnostics, createdId, data.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"running", "stopped"}, // In some cases, when there is insufficient capacity the VM is created with state = stopped
		r.provider.ApiClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", createdId, err))
		return
	}

	vmSchema, ok := read.(*iaas.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to create VM", "object conversion error")
		return
	}

	// In some cases, when there is insufficient capacity the VM is created with state = stopped
	if utils.GetPtrValue(vmSchema.State) == "stopped" {
		response.Diagnostics.AddError("Issue while creating VM", fmt.Sprintf("VM was created in 'stopped' state. Reason : %s", utils.GetPtrValue(vmSchema.StateReason)))
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
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadVmsByIdResponse, error) {
		id := utils.FromTfStringToStringPtr(data.Id)
		if id == nil {
			return nil, errors.New("Found invalid id")
		}
		return r.provider.ApiClient.ReadVmsByIdWithResponse(ctx, r.provider.SpaceID, *id)
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
	var state, plan resource_vm.VmModel

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
			r.provider.ApiClient,
			r.provider.SpaceID,
			vmId,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	body := VmFromTfToUpdaterequest(ctx, &plan, &response.Diagnostics)
	bodyFromState := VmFromTfToUpdaterequest(ctx, &state, &response.Diagnostics)

	if isUpdateNeeded(body, bodyFromState) {
		// Stop VM before doing update
		diags := StopVm(ctx, r.provider, vmId)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}

		// Update VM
		updatedRes := utils.ExecuteRequest(func() (*iaas.UpdateVmResponse, error) {
			return r.provider.ApiClient.UpdateVmWithResponse(ctx, r.provider.SpaceID, vmId, body)
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
	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		vmId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"running"},
		r.provider.ApiClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to update VM", fmt.Sprintf("Error waiting for VM to be created: %s", err))
		return
	}
	vmObject, ok := read.(*iaas.Vm)
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
	return utils.GetPtrValue(val1) == utils.GetPtrValue(val2)
}

func compareSlicePtr[R comparable](val1 *[]R, val2 *[]R) bool {
	return slices.Equal(utils.GetPtrValue(val1), utils.GetPtrValue(val2))
}

func isUpdateNeeded(plan iaas.UpdateVmJSONRequestBody, state iaas.UpdateVmJSONRequestBody) bool {
	return !(compareSimpleFieldPtr(plan.BsuOptimized, state.BsuOptimized) &&
		compareSimpleFieldPtr(plan.DeletionProtection, state.DeletionProtection) &&
		compareSimpleFieldPtr(plan.KeypairName, state.KeypairName) &&
		compareSimpleFieldPtr(plan.NestedVirtualization, state.NestedVirtualization) &&
		(utils.GetPtrValue(plan.Performance) == "" || compareSimpleFieldPtr(plan.Performance, state.Performance)) && // if performance is not provided by user,
		(len(utils.GetPtrValue(plan.BlockDeviceMappings)) == 0 || compareSlicePtr(plan.SecurityGroupIds, state.SecurityGroupIds)) &&
		compareSimpleFieldPtr(plan.UserData, state.UserData) &&
		compareSimpleFieldPtr(plan.VmInitiatedShutdownBehavior, state.VmInitiatedShutdownBehavior) &&
		compareSimpleFieldPtr(plan.Type, state.Type) &&
		(len(utils.GetPtrValue(plan.BlockDeviceMappings)) == 0 || (compareSlicePtr(plan.BlockDeviceMappings, state.BlockDeviceMappings))) &&
		compareSimpleFieldPtr(plan.IsSourceDestChecked, state.IsSourceDestChecked))
}

func (r *VmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VM", err.Error())
		return
	}
}
