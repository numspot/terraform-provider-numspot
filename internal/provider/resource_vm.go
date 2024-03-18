package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
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

	res := utils.ExecuteRequest(func() (*api.CreateVmsResponse, error) {
		body := VmFromTfToCreateRequest(ctx, &data)
		return r.provider.ApiClient.CreateVmsWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
		return
	}

	vm := *res.JSON201
	createdId := vm.Id

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"running"},
		Refresh: func() (result interface{}, state string, err error) {
			readed := r.readVmById(ctx, createdId, response.Diagnostics)
			if readed == nil {
				return nil, "", nil
			}

			return *readed, *readed.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", *createdId, err))
		return
	}

	vmSchema := read.(api.Vm)
	tf, diagnostics := VmFromHttpToTf(ctx, &vmSchema)
	if diagnostics.HasError() {
		response.Diagnostics.Append(diagnostics...)
		return
	}

	tf.Id = types.StringPointerValue(createdId)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) readVmById(ctx context.Context, id *string, diagnostics diag.Diagnostics) *api.Vm {
	res, err := r.provider.ApiClient.ReadVmsByIdWithResponse(ctx, r.provider.SpaceID, *id)
	if err != nil {
		diagnostics.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diagnostics.AddError("Failed to read Vm", apiError.Error())
		return nil
	}

	return res.JSON200
}

func (r *VmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadVmsByIdResponse, error) {
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

	res := utils.ExecuteRequest(func() (*api.DeleteVmsResponse, error) {
		idsSlice := make([]interface{}, 1)
		idsSlice[0] = data.Id.ValueString()
		return r.provider.ApiClient.DeleteVmsWithResponse(ctx, r.provider.SpaceID, idsSlice[0].(string))
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "running", "stopping", "shutting-down"},
		Target:  []string{"terminated"},
		Refresh: func() (result interface{}, state string, err error) {
			readed := r.readVmById(ctx, data.Id.ValueStringPointer(), response.Diagnostics)
			if readed == nil {
				return nil, "", nil
			}

			return *readed, *readed.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	_, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VM", fmt.Sprintf("Error waiting for instance (%s) to be deleted: %s", data.Id.ValueString(), err))
		return
	}
}
