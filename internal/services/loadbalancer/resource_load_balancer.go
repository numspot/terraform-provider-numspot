package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &LoadBalancerResource{}
	_ resource.ResourceWithConfigure   = &LoadBalancerResource{}
	_ resource.ResourceWithImportState = &LoadBalancerResource{}
)

type LoadBalancerResource struct {
	provider services.IProvider
}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

func (r *LoadBalancerResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(services.IProvider)
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
	response.Schema = LoadBalancerResourceSchema(ctx)
}

func (r *LoadBalancerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data LoadBalancerModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	// Retries create until request response is OK
	_, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.GetSpaceID(),
		LoadBalancerFromTfToCreateRequest(ctx, &data),
		r.provider.GetNumspotClient().CreateLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Load Balancer", err.Error())
		return
	}

	// Backends
	if len(data.BackendVmIds.Elements()) > 0 || len(data.BackendIps.Elements()) > 0 {
		diags := r.linkBackendMachines(ctx,
			data.Name.ValueString(),
			utils.TfStringSetToStringPtrSet(ctx, data.BackendIps),
			utils.TfStringSetToStringPtrSet(ctx, data.BackendVmIds),
		)
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
		CreateLoadBalancerTags(ctx, r.provider.GetSpaceID(), r.provider.GetNumspotClient(), data.Name.ValueString(), data.Tags, &response.Diagnostics)
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

func (r *LoadBalancerResource) readLoadBalancer(ctx context.Context, id string, diagnostic diag.Diagnostics) *LoadBalancerModel {
	res := utils.ExecuteRequest(func() (*numspot.ReadLoadBalancersByIdResponse, error) {
		return r.provider.GetNumspotClient().ReadLoadBalancersByIdWithResponse(ctx, r.provider.GetSpaceID(), id)
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
	var data LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	tf := r.readLoadBalancer(ctx, data.Id.ValueString(), response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan LoadBalancerModel
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

		DeleteLoadBalancerTags(ctx, r.provider.GetSpaceID(), r.provider.GetNumspotClient(), state.Name.ValueString(), toDeleteTf, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		toCreateTf, diags = types.ListValueFrom(ctx, tagType, toCreate)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		CreateLoadBalancerTags(ctx, r.provider.GetSpaceID(), r.provider.GetNumspotClient(), state.Name.ValueString(), toCreateTf, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		modifs = true
	}

	if !plan.Listeners.IsUnknown() && !plan.Listeners.Equal(state.Listeners) {
		stateListeners := make([]ListenersValue, 0, len(state.Listeners.Elements()))
		planListeners := make([]ListenersValue, 0, len(plan.Listeners.Elements()))

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

	if (!plan.BackendIps.IsUnknown() && !plan.BackendIps.Equal(state.BackendIps)) ||
		(!plan.BackendVmIds.IsUnknown() && !plan.BackendVmIds.Equal(state.BackendVmIds)) {
		diags := r.updateBackendMachines(ctx, plan, state)
		if diags.HasError() {
			response.Diagnostics.Append(diags...)
		}
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

func (r *LoadBalancerResource) createListeners(ctx context.Context, loadBalancerId string, listeners []ListenersValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	listenersForCreation := make([]numspot.ListenerForCreation, 0, len(listeners))
	for _, e := range listeners {
		listenersForCreation = append(listenersForCreation, numspot.ListenerForCreation{
			BackendPort:          utils.FromTfInt64ToInt(e.BackendPort),
			BackendProtocol:      utils.FromTfStringToStringPtr(e.BackendProtocol),
			LoadBalancerPort:     utils.FromTfInt64ToInt(e.LoadBalancerPort),
			LoadBalancerProtocol: e.LoadBalancerProtocol.ValueString(),
		})
	}

	utils.ExecuteRequest(func() (*numspot.CreateLoadBalancerListenersResponse, error) {
		return r.provider.GetNumspotClient().CreateLoadBalancerListenersWithResponse(
			ctx,
			r.provider.GetSpaceID(),
			loadBalancerId,
			numspot.CreateLoadBalancerListenersJSONRequestBody{
				Listeners: listenersForCreation,
			},
		)
	}, http.StatusCreated, &diags)

	return diags
}

func (r *LoadBalancerResource) deleteListeners(ctx context.Context, loadBalancerId string, listeners []ListenersValue) diag.Diagnostics {
	diags := diag.Diagnostics{}

	listenersLoadBalancerPortToDelete := make([]int, 0, len(listeners))
	for _, e := range listeners {
		listenersLoadBalancerPortToDelete = append(listenersLoadBalancerPortToDelete, utils.FromTfInt64ToInt(e.LoadBalancerPort))
	}

	utils.ExecuteRequest(func() (*numspot.DeleteLoadBalancerListenersResponse, error) {
		return r.provider.GetNumspotClient().DeleteLoadBalancerListenersWithResponse(
			ctx,
			r.provider.GetSpaceID(),
			loadBalancerId,
			numspot.DeleteLoadBalancerListenersJSONRequestBody{
				LoadBalancerPorts: listenersLoadBalancerPortToDelete,
			},
		)
	}, http.StatusNoContent, &diags)

	return diags
}

func (r *LoadBalancerResource) shouldUpdate(state, plan LoadBalancerModel) bool {
	shouldUpdate := false

	shouldUpdate = shouldUpdate || (!plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck))
	shouldUpdate = shouldUpdate || (!plan.SecurityGroups.IsUnknown() && !state.SecurityGroups.Equal(plan.SecurityGroups))
	shouldUpdate = shouldUpdate || (!plan.SecuredCookies.IsUnknown() && !state.SecuredCookies.Equal(plan.SecuredCookies))

	return shouldUpdate
}

// Outscale only allows one update at a time for loadbalancer (don't ask why), so we call the update function multiple time
func (r *LoadBalancerResource) UpdateLoadBalancer(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var plan, state LoadBalancerModel

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
		res := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
			return r.provider.GetNumspotClient().UpdateLoadBalancerWithResponse(ctx, r.provider.GetSpaceID(), plan.Name.ValueString(), numspot.UpdateLoadBalancerJSONRequestBody{
				HealthCheck: &numspot.HealthCheck{
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
		res := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
			securityGroups := utils.FromTfStringListToStringList(ctx, plan.SecurityGroups)
			return r.provider.GetNumspotClient().UpdateLoadBalancerWithResponse(ctx, r.provider.GetSpaceID(), plan.Name.ValueString(), numspot.UpdateLoadBalancerJSONRequestBody{
				SecurityGroups: &securityGroups,
			})
		}, http.StatusOK, &response.Diagnostics)
		if res == nil {
			return
		}
	}
}

func (r *LoadBalancerResource) updateBackendMachines(ctx context.Context, plan, state LoadBalancerModel) diag.Diagnostics {
	var diags diag.Diagnostics
	var backendIpstoLink, backendIpstoUnlink, backendVmIdstoLink, backendVmIdstoUnlink []string

	stateBackendIps := utils.TfStringSetToStringPtrSet(ctx, state.BackendIps)
	planBackendIps := utils.TfStringSetToStringPtrSet(ctx, plan.BackendIps)
	stateBackendVmIds := utils.TfStringSetToStringPtrSet(ctx, state.BackendVmIds)
	planBackendVmIds := utils.TfStringSetToStringPtrSet(ctx, plan.BackendVmIds)

	if planBackendIps != nil && stateBackendIps != nil {
		backendIpstoLink, backendIpstoUnlink = utils.DiffComparable(*planBackendIps, *stateBackendIps)
	}
	if planBackendIps != nil && stateBackendIps != nil {
		backendVmIdstoLink, backendVmIdstoUnlink = utils.DiffComparable(*planBackendVmIds, *stateBackendVmIds)
	}

	diags = r.unlinkBackendMachines(ctx, state.Name.ValueString(), &backendIpstoUnlink, &backendVmIdstoUnlink)
	if diags.HasError() {
		return diags
	}

	diags = r.linkBackendMachines(ctx, state.Name.ValueString(), &backendIpstoLink, &backendVmIdstoLink)
	if diags.HasError() {
		return diags
	}

	return nil
}

func (r *LoadBalancerResource) unlinkBackendMachines(ctx context.Context, lbName string, backendIps, backendVmIds *[]string) diag.Diagnostics {
	var diags diag.Diagnostics

	payload := numspot.UnlinkLoadBalancerBackendMachinesJSONRequestBody{}

	if backendIps != nil && len(*backendIps) > 0 {
		payload.BackendIps = backendIps
	}

	if backendVmIds != nil && len(*backendVmIds) > 0 {
		payload.BackendVmIds = backendVmIds
	}

	utils.ExecuteRequest(func() (*numspot.UnlinkLoadBalancerBackendMachinesResponse, error) {
		return r.provider.GetNumspotClient().UnlinkLoadBalancerBackendMachinesWithResponse(ctx, r.provider.GetSpaceID(), lbName, payload)
	}, http.StatusNoContent, &diags)
	return diags
}

func (r *LoadBalancerResource) linkBackendMachines(ctx context.Context, lbName string, backendIps, backendVmIds *[]string) diag.Diagnostics {
	var diags diag.Diagnostics

	payload := numspot.LinkLoadBalancerBackendMachinesJSONRequestBody{}

	if backendIps != nil && len(*backendIps) > 0 {
		payload.BackendIps = backendIps
	}

	if backendVmIds != nil && len(*backendVmIds) > 0 {
		payload.BackendVmIds = backendVmIds
	}

	utils.ExecuteRequest(func() (*numspot.LinkLoadBalancerBackendMachinesResponse, error) {
		return r.provider.GetNumspotClient().LinkLoadBalancerBackendMachinesWithResponse(ctx, r.provider.GetSpaceID(), lbName, payload)
	}, http.StatusNoContent, &diags)
	return diags
}

func (r *LoadBalancerResource) AttachHealthCheck(ctx context.Context, lbName string, healthCheck HealthCheckValue) (*numspot.UpdateLoadBalancerResponse, diag.Diagnostics) {
	var diags diag.Diagnostics

	payload := numspot.UpdateLoadBalancerJSONRequestBody{
		HealthCheck: &numspot.HealthCheck{
			CheckInterval:      utils.FromTfInt64ToInt(healthCheck.CheckInterval),
			HealthyThreshold:   utils.FromTfInt64ToInt(healthCheck.HealthyThreshold),
			Path:               healthCheck.Path.ValueStringPointer(),
			Port:               utils.FromTfInt64ToInt(healthCheck.Port),
			Protocol:           healthCheck.Protocol.ValueString(),
			Timeout:            utils.FromTfInt64ToInt(healthCheck.Timeout),
			UnhealthyThreshold: utils.FromTfInt64ToInt(healthCheck.UnhealthyThreshold),
		},
	}

	updatedLoadBalancer := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
		return r.provider.GetNumspotClient().UpdateLoadBalancerWithResponse(ctx, r.provider.GetSpaceID(), lbName, payload)
	}, http.StatusOK, &diags)

	return updatedLoadBalancer, diags
}

func (r *LoadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	tflog.Debug(ctx, fmt.Sprintf("Deleting load balancer %s", data.Id))

	// Detach security groups
	emptyList := []string{}
	res := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
		return r.provider.GetNumspotClient().UpdateLoadBalancerWithResponse(ctx, r.provider.GetSpaceID(), data.Name.ValueString(), numspot.UpdateLoadBalancerJSONRequestBody{
			SecurityGroups: &emptyList,
		})
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	err := utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.GetSpaceID(), data.Id.ValueString(), r.provider.GetNumspotClient().DeleteLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Load Balancer", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
