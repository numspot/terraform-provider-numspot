package provider

import (
	"context"
	"fmt"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
)

var (
	_ resource.Resource                = &LoadBalancerResource{}
	_ resource.ResourceWithConfigure   = &LoadBalancerResource{}
	_ resource.ResourceWithImportState = &LoadBalancerResource{}
)

type LoadBalancerResource struct {
	client *api.ClientWithResponses
}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

func (r *LoadBalancerResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
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

	body := LoadBalancerFromTfToCreateRequest(ctx, &data)
	createRes := utils.ExecuteRequest(func() (*api.CreateLoadBalancerResponse, error) {
		return r.client.CreateLoadBalancerWithResponse(ctx, body)
	}, http.StatusOK, &response.Diagnostics)
	if createRes == nil {
		return
	}

	tf := LoadBalancerFromHttpToTf(ctx, createRes.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.ReadLoadBalancersByIdResponse, error) {
		return r.client.ReadLoadBalancersByIdWithResponse(ctx, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf := LoadBalancerFromHttpToTf(ctx, res.JSON200) // FIXME
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	// TODO implement me
	panic("implement me")
}

func (r *LoadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res := utils.ExecuteRequest(func() (*api.DeleteLoadBalancerResponse, error) {
		return r.client.DeleteLoadBalancerWithResponse(ctx, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
}
