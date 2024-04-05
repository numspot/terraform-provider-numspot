package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
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

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		VmFromTfToCreateRequest(ctx, &data),
		r.provider.ApiClient.CreateVmsWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", err.Error())
		return
	}

	vm := *res.JSON201
	createdId := vm.Id
	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		*createdId,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"running"},
		r.provider.ApiClient.ReadVmsByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", *createdId, err))
		return
	}

	vmSchema, ok := read.(iaas.Vm)
	if !ok {
		response.Diagnostics.AddError("Failed to create VM", "object conversion error")
	}
	tf, diagnostics := VmFromHttpToTf(ctx, &vmSchema)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	tf.Id = types.StringPointerValue(createdId)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadVmsByIdResponse, error) {
		id := data.Id.ValueStringPointer()
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
	panic("implement me")
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
