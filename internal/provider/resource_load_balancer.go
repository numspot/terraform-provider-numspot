package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
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
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	_, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		LoadBalancerFromTfToCreateRequest(ctx, &data),
		r.provider.ApiClient.CreateLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Load Balancer", err.Error())
		return
	}

	// Backends
	if len(data.BackendVmIds.Elements()) > 0 || len(data.BackendIps.Elements()) > 0 {
		diags := r.linkBackendMachines(ctx, data)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	// Health Check
	if !data.HealthCheck.IsUnknown() {
		_, diags := r.AttachHealthCheck(ctx, data.Name.ValueString(), data.HealthCheck)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
			return
		}
	}

	// Tags
	// In load balancer case, tags are not handled in the same way as other resources
	if len(data.Tags.Elements()) > 0 {
		CreateLoadBalancerTags(ctx, r.provider.SpaceID, r.provider.ApiClient, data.Name.ValueString(), data.Tags, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Update state with updated resource
	res := utils.ExecuteRequest(func() (*iaas.ReadLoadBalancersByIdResponse, error) {
		return r.provider.ApiClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, data.Name.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := LoadBalancerFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	res := utils.ExecuteRequest(func() (*iaas.ReadLoadBalancersByIdResponse, error) {
		return r.provider.ApiClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, data.Id.ValueString())
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := LoadBalancerFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	r.UpdateLoadBalancer(ctx, request, response)
}

func (r *LoadBalancerResource) UpdateLoadBalancer(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := LoadBalancerFromTfToUpdateRequest(ctx, &plan)

	res := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
		return r.provider.ApiClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), payload)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	tf, diags := LoadBalancerFromHttpToTf(ctx, res.JSON200)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *LoadBalancerResource) linkBackendMachines(ctx context.Context, plan resource_load_balancer.LoadBalancerModel) diag.Diagnostics {
	var diags diag.Diagnostics

	payload := iaas.LinkLoadBalancerBackendMachinesJSONRequestBody{}
	if !plan.BackendIps.IsUnknown() {
		payload.BackendIps = utils.TfStringSetToStringPtrSet(ctx, plan.BackendIps)
	}
	if !plan.BackendVmIds.IsUnknown() {
		payload.BackendVmIds = utils.TfStringSetToStringPtrSet(ctx, plan.BackendVmIds)
	}

	_ = utils.ExecuteRequest(func() (*iaas.LinkLoadBalancerBackendMachinesResponse, error) {
		return r.provider.ApiClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), payload)
	}, http.StatusNoContent, &diags)
	return diags
}

func (r *LoadBalancerResource) AttachHealthCheck(ctx context.Context, lbName string, healthCheck resource_load_balancer.HealthCheckValue) (*iaas.UpdateLoadBalancerResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := LoadBalancerHealthCheckToAttachHealthCheckRequest(healthCheck)

	updatedLoadBalancer := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
		return r.provider.ApiClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, lbName, payload)
	}, http.StatusOK, &diags)

	return updatedLoadBalancer, diags
}

func (r *LoadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Deleting load balancer %s", data.Id))

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.ApiClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Load Balancer", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
