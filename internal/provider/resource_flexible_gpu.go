package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_flexible_gpu"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &FlexibleGpuResource{}
	_ resource.ResourceWithConfigure   = &FlexibleGpuResource{}
	_ resource.ResourceWithImportState = &FlexibleGpuResource{}
)

type FlexibleGpuResource struct {
	provider Provider
}

func NewFlexibleGpuResource() resource.Resource {
	return &FlexibleGpuResource{}
}

func (r *FlexibleGpuResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		FlexibleGpuFromTfToCreateRequest(&data),
		r.provider.ApiClient.CreateFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Flexible GPU", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		createdId,
		r.provider.SpaceID,
		[]string{"attaching", "detaching"},
		[]string{"allocated", "attached"},
		r.provider.ApiClient.ReadFlexibleGpusByIdWithResponse,
	)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Flexible GPU", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", data.Id.ValueString(), err))
		return
	}

	flexGPU, ok := read.(*iaas.FlexibleGpu)
	if !ok {
		response.Diagnostics.AddError("Failed to create Flexible GPU", "object conversion error")
		return
	}
	tf := FlexibleGpuFromHttpToTf(flexGPU)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *FlexibleGpuResource) read(ctx context.Context, id string, diagnostics diag.Diagnostics) *iaas.FlexibleGpu {
	res, err := r.provider.ApiClient.ReadFlexibleGpusByIdWithResponse(ctx, r.provider.SpaceID, id)
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
	var plan, state resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if plan.DeleteOnVmDeletion != state.DeleteOnVmDeletion {
		body := FlexibleGpuFromTfToUpdateRequest(&plan)
		res := utils.ExecuteRequest(func() (*iaas.UpdateFlexibleGpuResponse, error) {
			return r.provider.ApiClient.UpdateFlexibleGpuWithResponse(
				ctx,
				r.provider.SpaceID,
				state.Id.ValueString(),
				body)
		}, http.StatusOK, &response.Diagnostics)
		if res == nil {
			return
		}

		state = FlexibleGpuFromHttpToTf(res.JSON200)
		response.Diagnostics.Append(response.State.Set(ctx, state)...)
	}
}

func (r *FlexibleGpuResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteFlexibleGpuWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Flexible GPU", err.Error())
		return
	}
}
