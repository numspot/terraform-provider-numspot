package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
)

var (
	_ resource.Resource                = &FlexibleGpuResource{}
	_ resource.ResourceWithConfigure   = &FlexibleGpuResource{}
	_ resource.ResourceWithImportState = &FlexibleGpuResource{}
)

type FlexibleGpuResource struct {
	client *api.ClientWithResponses
}

func NewFlexibleGpuResource() resource.Resource {
	return &FlexibleGpuResource{}
}

func (r *FlexibleGpuResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

func (r *FlexibleGpuResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *FlexibleGpuResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_flexible_gpu"
}

func (r *FlexibleGpuResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_flexible_gpu.FlexibleGpuResourceSchema(ctx)
}

func (r *FlexibleGpuResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	res := utils.HandleResponse(func() (*api.CreateFlexibleGpuResponse, error) {
		body := FlexibleGpuFromTfToCreateRequest(&data)
		return r.client.CreateFlexibleGpuWithResponse(ctx, body)
	}, http.StatusOK, &response.Diagnostics)

	createStateConf := &retry.StateChangeConf{
		Pending: []string{"attaching", "detaching"},
		Target:  []string{"allocated", "attached"},
		Refresh: func() (result interface{}, state string, err error) {
			readed := r.read(ctx, *res.JSON200.Id, response.Diagnostics)
			if readed == nil {
				return nil, "", nil
			}

			return *readed, *readed.State, nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	read, err := createStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VM", fmt.Sprintf("Error waiting for example instance (%s) to be created: %s", data.Id.ValueString(), err))
		return
	}

	tf := FlexibleGpuFromHttpToTf(read.(*api.FlexibleGpuSchema))
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) read(ctx context.Context, id string, diagnostics diag.Diagnostics) *api.FlexibleGpuSchema {
	res, err := r.client.ReadFlexibleGpusByIdWithResponse(ctx, id)
	if err != nil {
		diagnostics.AddError("Failed to read RouteTable", err.Error())
		return nil
	}

	if res.StatusCode() != http.StatusOK {
		apiError := utils.HandleError(res.Body)
		diagnostics.AddError("Failed to read FlexibleGpu", apiError.Error())
		return nil
	}

	return res.JSON200
}

func (r *FlexibleGpuResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	gpu := r.read(ctx, data.Id.ValueString(), response.Diagnostics)
	if gpu == nil {
		return
	}

	tf := FlexibleGpuFromHttpToTf(gpu)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	panic("implement me")
}

func (r *FlexibleGpuResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	utils.HandleResponse(func() (*api.DeleteFlexibleGpuResponse, error) {
		return r.client.DeleteFlexibleGpuWithResponse(ctx, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
}
