package provider

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/iaas"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
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

	body := LoadBalancerFromTfToCreateRequest(ctx, &data)
	res := utils.ExecuteRequest(func() (*iaas.CreateLoadBalancerResponse, error) {
		return r.provider.ApiClient.CreateLoadBalancerWithResponse(ctx, r.provider.SpaceID, body)
	}, http.StatusCreated, &response.Diagnostics)
	if res == nil {
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

	res := utils.ExecuteRequest(func() (*iaas.DeleteLoadBalancerResponse, error) {
		return r.provider.ApiClient.DeleteLoadBalancerWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusNoContent, &response.Diagnostics)
	if res == nil {
		return
	}

	deleteStateConf := &retry.StateChangeConf{
		Pending: []string{"pending", "available", "deleting"},
		Target:  []string{"deleted"},
		Refresh: func() (result interface{}, state string, err error) {
			// Do not use utils.ExecuteRequest to access error response to know if it's a 404 Not Found expected response
			readLbRes, err := r.provider.ApiClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
			if err != nil {
				response.Diagnostics.AddError("Failed to read Net on delete", err.Error())
				return
			}

			if readLbRes.StatusCode() != http.StatusOK {
				if readLbRes.StatusCode() == http.StatusNotFound {
					return data, "deleted", nil
				}
				apiError := utils.HandleError(readLbRes.Body)
				response.Diagnostics.AddError("Failed to read load balancer on delete", apiError.Error())
				return
			}

			return data, "pending", nil
		},
		Timeout: 5 * time.Minute,
		Delay:   5 * time.Second,
	}

	_, err := deleteStateConf.WaitForStateContext(ctx)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Net", fmt.Sprintf("Error waiting for instance (%s) to be deleted: %s", data.Id.ValueString(), err))
		return
	}

	response.State.RemoveResource(ctx)
}
