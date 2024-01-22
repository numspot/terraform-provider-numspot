package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vm"
)

var _ resource.Resource = &VmResource{}
var _ resource.ResourceWithConfigure = &VmResource{}
var _ resource.ResourceWithImportState = &VmResource{}

type VmResource struct {
	client *api.ClientWithResponses
}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

func (r *VmResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*api.ClientWithResponses)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.client = client
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

	body := VmFromTfToCreateRequest(data)
	res, err := r.client.CreateVmsWithResponse(ctx, body)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to create Vm", err.Error())
		return
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 201)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to create Vm", apiError.Error())
		return
	}

	if res.JSON200 == nil || res.JSON200.Vms == nil {
		response.Diagnostics.AddError("Failed to create Vm", "My Custom Error")
		return
	}
	vms := *res.JSON200.Vms
	createdId := vms[0].Id

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"running"},
		Refresh: func() (result interface{}, state string, err error) {
			res, err := r.client.ReadVmsByIdWithResponse(ctx, *createdId)
			if err != nil {
				// TODO: Handle Error
				response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
			}

			expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
			if res.StatusCode() != expectedStatusCode {
				apiError := utils.HandleError(res.Body)
				response.Diagnostics.AddError("Failed to read Vm", apiError.Error())
				return
			}

			return *res.JSON200, *res.JSON200.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   3 * time.Second,
	}

	rr, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", *createdId, err))
		return
	}

	vmSchema := rr.(api.VmSchema)
	tf := VmFromHttpToTf(&vmSchema)
	tf.Id = types.StringPointerValue(createdId)

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	id := data.Id.ValueString()
	res, err := r.client.ReadVmsByIdWithResponse(ctx, id)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to read RouteTable", err.Error())
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 200)
	if res.StatusCode() != expectedStatusCode {
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to read Vm", apiError.Error())
		return
	}

	tf := VmFromHttpToTf(res.JSON200) // FIXME
	tf.Id = types.StringValue(id)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *VmResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *VmResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_vm.VmModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	//TODO: Implement DELETE operation
	interfaceSlice := make([]interface{}, 1)
	interfaceSlice[0] = data.Id.ValueString()
	res, err := r.client.DeleteVmsWithResponse(ctx, interfaceSlice)
	if err != nil {
		// TODO: Handle Error
		response.Diagnostics.AddError("Failed to delete Vm", err.Error())
		return
	}

	expectedStatusCode := 200 //FIXME: Set expected status code (must be 204)
	if res.StatusCode() != expectedStatusCode {
		// TODO: Handle NumSpot error`
		apiError := utils.HandleError(res.Body)
		response.Diagnostics.AddError("Failed to delete Vm", apiError.Error())
		return
	}
}
