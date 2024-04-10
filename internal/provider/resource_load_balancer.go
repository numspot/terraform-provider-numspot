package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &LoadBalancerResource{}
	_ resource.ResourceWithConfigure   = &LoadBalancerResource{}
	_ resource.ResourceWithImportState = &LoadBalancerResource{}
)

type LoadBalancerResource struct {
	provider Provider
}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

func (r *LoadBalancerResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = client
}

func (r *LoadBalancerResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *LoadBalancerResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_load_balancer"
}

func (r *LoadBalancerResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_load_balancer.LoadBalancerResourceSchema(ctx)
}

func (r *LoadBalancerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		LoadBalancerFromTfToCreateRequest(ctx, &data),
		r.provider.ApiClient.CreateLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Load Balancer", err.Error())
		return
	}

	tf := LoadBalancerFromHttpToTf(ctx, res.JSON201)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*iaas.ReadLoadBalancersByIdResponse, error) {
		return r.provider.ApiClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := LoadBalancerFromHttpToTf(ctx, res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	if !plan.BackendIps.IsUnknown() || !plan.BackendVmIds.IsUnknown() {
		r.LinkBackendMachines(ctx, request, response)
		return
	}
	r.UpdateLoadBalancer(ctx, request, response)
}

func (r *LoadBalancerResource) UpdateLoadBalancer(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)

	payload := LoadBalancerFromTfToUpdateRequest(ctx, &plan)

	res := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
		return r.provider.ApiClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), payload)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	tf := LoadBalancerFromHttpToTf(ctx, res.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *LoadBalancerResource) LinkBackendMachines(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_load_balancer.LoadBalancerModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)

	payload := iaas.LinkLoadBalancerBackendMachinesJSONRequestBody{}
	if !plan.BackendIps.IsUnknown() {
		payload.BackendIps = utils.TfStringListToStringPtrList(ctx, plan.BackendIps)
	}
	if !plan.BackendVmIds.IsUnknown() {
		payload.BackendVmIds = utils.TfStringListToStringPtrList(ctx, plan.BackendVmIds)
	}

	res := utils.ExecuteRequest(func() (*iaas.LinkLoadBalancerBackendMachinesResponse, error) {
		return r.provider.ApiClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), payload)
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	resRead := utils.ExecuteRequest(func() (*iaas.ReadLoadBalancersByIdResponse, error) {
		return r.provider.ApiClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, state.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if resRead == nil {
		return
	}

	tf := LoadBalancerFromHttpToTf(ctx, resRead.JSON200)
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *LoadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Load Balancer", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
