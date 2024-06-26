package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_load_balancer"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
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
	_, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		LoadBalancerFromTfToCreateRequest(ctx, &data),
		r.provider.IaasClient.CreateLoadBalancerWithResponse)
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
		CreateLoadBalancerTags(ctx, r.provider.SpaceID, r.provider.IaasClient, data.Name.ValueString(), data.Tags, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Update state with updated resource
	tf := r.readLoadBalancer(ctx, data.Name.ValueString(), response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) readLoadBalancer(ctx context.Context, id string, diagnostic diag.Diagnostics) *resource_load_balancer.LoadBalancerModel {
	res := utils.ExecuteRequest(func() (*iaas.ReadLoadBalancersByIdResponse, error) {
		return r.provider.IaasClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, id)
	}, http.StatusOK, &diagnostic)
	if res == nil {
		return nil
	}

	tf, diags := LoadBalancerFromHttpToTf(ctx, res.JSON200)
	diagnostic.Append(diags...)
	if diagnostic.HasError() {
		return nil
	}

	return tf
}

func (r *LoadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	tf := r.readLoadBalancer(ctx, data.Id.ValueString(), response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan resource_load_balancer.LoadBalancerModel
		diags       diag.Diagnostics
		modifs      = false
	)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Tags.IsUnknown() && !plan.Tags.Equal(state.Tags) {
		stateTags := make([]tags.TagsValue, 0, len(state.Tags.Elements()))
		diags = state.Tags.ElementsAs(ctx, &stateTags, false)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		planTags := make([]tags.TagsValue, 0, len(plan.Tags.Elements()))
		diags = plan.Tags.ElementsAs(ctx, &planTags, false)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		toCreate, toDelete := utils.Diff(stateTags, planTags)
		var toDeleteTf, toCreateTf types.List
		tagType := tags.TagsValue{}.Type(ctx)

		toDeleteTf, diags = types.ListValueFrom(ctx, tagType, toDelete)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		DeleteLoadBalancerTags(ctx, r.provider.SpaceID, r.provider.IaasClient, state.Name.ValueString(), toDeleteTf, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		toCreateTf, diags = types.ListValueFrom(ctx, tagType, toCreate)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		CreateLoadBalancerTags(ctx, r.provider.SpaceID, r.provider.IaasClient, state.Name.ValueString(), toCreateTf, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		modifs = true
	}

	if !plan.Listeners.IsUnknown() && !plan.Listeners.Equal(state.Listeners) {
		stateListeners := make([]resource_load_balancer.ListenersValue, 0, len(state.Listeners.Elements()))
		planListeners := make([]resource_load_balancer.ListenersValue, 0, len(plan.Listeners.Elements()))

		diags = state.Listeners.ElementsAs(ctx, &stateListeners, false)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		diags = plan.Listeners.ElementsAs(ctx, &planListeners, false)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		toCreate, toDelete := utils.Diff(stateListeners, planListeners)

		if len(toCreate) > 0 {
			response.Diagnostics.Append(r.createListeners(ctx, state.Id.ValueString(), toCreate)...)
			if response.Diagnostics.HasError() {
				return
			}
		}

		if len(toDelete) > 0 {
			response.Diagnostics.Append(r.deleteListeners(ctx, state.Id.ValueString(), toDelete)...)
			if response.Diagnostics.HasError() {
				return
			}
		}

		modifs = true
	}

	if r.shouldUpdate(state, plan) {
		r.UpdateLoadBalancer(ctx, request, response)
		if response.Diagnostics.HasError() {
			return
		}

		modifs = true
	}

	if modifs {
		tf := r.readLoadBalancer(ctx, state.Id.ValueString(), response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
	}
}

func (r *LoadBalancerResource) createListeners(ctx context.Context, loadBalancerId string, listeners []resource_load_balancer.ListenersValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	listenersForCreation := make([]iaas.ListenerForCreation, 0, len(listeners))
	for _, e := range listeners {
		listenersForCreation = append(listenersForCreation, iaas.ListenerForCreation{
			BackendPort:          utils.FromTfInt64ToInt(e.BackendPort),
			BackendProtocol:      utils.FromTfStringToStringPtr(e.BackendProtocol),
			LoadBalancerPort:     utils.FromTfInt64ToInt(e.LoadBalancerPort),
			LoadBalancerProtocol: e.LoadBalancerProtocol.ValueString(),
		})
	}

	utils.ExecuteRequest(func() (*iaas.CreateLoadBalancerListenersResponse, error) {
		return r.provider.IaasClient.CreateLoadBalancerListenersWithResponse(
			ctx,
			r.provider.SpaceID,
			loadBalancerId,
			iaas.CreateLoadBalancerListenersJSONRequestBody{
				Listeners: listenersForCreation,
			},
		)
	}, http.StatusCreated, &diags)

	return diags
}

func (r *LoadBalancerResource) deleteListeners(ctx context.Context, loadBalancerId string, listeners []resource_load_balancer.ListenersValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	listenersLoadBalancerPortToDelete := make([]int, 0, len(listeners))
	for _, e := range listeners {
		listenersLoadBalancerPortToDelete = append(listenersLoadBalancerPortToDelete, utils.FromTfInt64ToInt(e.LoadBalancerPort))
	}

	utils.ExecuteRequest(func() (*iaas.DeleteLoadBalancerListenersResponse, error) {
		return r.provider.IaasClient.DeleteLoadBalancerListenersWithResponse(
			ctx,
			r.provider.SpaceID,
			loadBalancerId,
			iaas.DeleteLoadBalancerListenersJSONRequestBody{
				LoadBalancerPorts: listenersLoadBalancerPortToDelete,
			},
		)
	}, http.StatusNoContent, &diags)

	return diags
}

func (r *LoadBalancerResource) shouldUpdate(state, plan resource_load_balancer.LoadBalancerModel) bool {
	shouldUpdate := false

	shouldUpdate = shouldUpdate || (!plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck))
	shouldUpdate = shouldUpdate || (!plan.SecurityGroups.IsUnknown() && !state.SecurityGroups.Equal(plan.SecurityGroups))
	shouldUpdate = shouldUpdate || (!plan.SecuredCookies.IsUnknown() && !state.SecuredCookies.Equal(plan.SecuredCookies))

	return shouldUpdate
}

func (r *LoadBalancerResource) UpdateLoadBalancer(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state resource_load_balancer.LoadBalancerModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	// HealthCheck:
	if !plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck) {
		res := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
			return r.provider.IaasClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), iaas.UpdateLoadBalancerJSONRequestBody{
				HealthCheck: &iaas.HealthCheck{
					CheckInterval:      utils.FromTfInt64ToInt(plan.HealthCheck.CheckInterval),
					HealthyThreshold:   utils.FromTfInt64ToInt(plan.HealthCheck.HealthyThreshold),
					Path:               plan.HealthCheck.Path.ValueStringPointer(),
					Port:               utils.FromTfInt64ToInt(plan.HealthCheck.Port),
					Protocol:           plan.HealthCheck.Protocol.ValueString(),
					Timeout:            utils.FromTfInt64ToInt(plan.HealthCheck.Timeout),
					UnhealthyThreshold: utils.FromTfInt64ToInt(plan.HealthCheck.UnhealthyThreshold),
				},
			})
		}, http.StatusOK, &response.Diagnostics)
		if res == nil {
			return
		}
	}

	// Security Groups
	if !plan.SecurityGroups.IsUnknown() && !state.SecurityGroups.Equal(plan.SecurityGroups) {
		res := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
			securityGroups := utils.FromTfStringListToStringList(ctx, plan.SecurityGroups)
			return r.provider.IaasClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), iaas.UpdateLoadBalancerJSONRequestBody{
				SecurityGroups: &securityGroups,
			})
		}, http.StatusOK, &response.Diagnostics)
		if res == nil {
			return
		}
	}
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

	utils.ExecuteRequest(func() (*iaas.LinkLoadBalancerBackendMachinesResponse, error) {
		return r.provider.IaasClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), payload)
	}, http.StatusNoContent, &diags)
	return diags
}

func (r *LoadBalancerResource) AttachHealthCheck(ctx context.Context, lbName string, healthCheck resource_load_balancer.HealthCheckValue) (*iaas.UpdateLoadBalancerResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := iaas.UpdateLoadBalancerJSONRequestBody{
		HealthCheck: &iaas.HealthCheck{
			CheckInterval:      utils.FromTfInt64ToInt(healthCheck.CheckInterval),
			HealthyThreshold:   utils.FromTfInt64ToInt(healthCheck.HealthyThreshold),
			Path:               healthCheck.Path.ValueStringPointer(),
			Port:               utils.FromTfInt64ToInt(healthCheck.Port),
			Protocol:           healthCheck.Protocol.ValueString(),
			Timeout:            utils.FromTfInt64ToInt(healthCheck.Timeout),
			UnhealthyThreshold: utils.FromTfInt64ToInt(healthCheck.UnhealthyThreshold),
		},
	}

	updatedLoadBalancer := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
		return r.provider.IaasClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, lbName, payload)
	}, http.StatusOK, &diags)

	return updatedLoadBalancer, diags
}

func (r *LoadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_load_balancer.LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	tflog.Debug(ctx, fmt.Sprintf("Deleting load balancer %s", data.Id))

	// Detach security groups
	emptyList := []string{}
	res := utils.ExecuteRequest(func() (*iaas.UpdateLoadBalancerResponse, error) {
		return r.provider.IaasClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, data.Name.ValueString(), iaas.UpdateLoadBalancerJSONRequestBody{
			SecurityGroups: &emptyList,
		})
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.IaasClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Load Balancer", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
