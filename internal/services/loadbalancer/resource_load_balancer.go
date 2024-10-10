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

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &LoadBalancerResource{}
	_ resource.ResourceWithConfigure   = &LoadBalancerResource{}
	_ resource.ResourceWithImportState = &LoadBalancerResource{}
)

type LoadBalancerResource struct {
	provider *client.NumSpotSDK
}

func NewLoadBalancerResource() resource.Resource {
	return &LoadBalancerResource{}
}

func (r *LoadBalancerResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(*client.NumSpotSDK)
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

	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	// Retries create until request response is OK
	_, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		LoadBalancerFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
		numspotClient.CreateLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create Load Balancer", err.Error())
		return
	}

	// Backends
	if len(data.BackendVmIds.Elements()) > 0 || len(data.BackendIps.Elements()) > 0 {
		r.linkBackendMachines(ctx,
			data.Name.ValueString(),
			utils.TfStringSetToStringPtrSet(ctx, data.BackendIps, &response.Diagnostics),
			utils.TfStringSetToStringPtrSet(ctx, data.BackendVmIds, &response.Diagnostics),
			&response.Diagnostics,
		)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Health Check
	if !data.HealthCheck.IsUnknown() {
		_ = r.AttachHealthCheck(ctx, data.Name.ValueString(), data.HealthCheck, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	// Tags
	// In load balancer case, tags are not handled in the same way as other resources
	if len(data.Tags.Elements()) > 0 {
		CreateLoadBalancerTags(ctx, r.provider.SpaceID, numspotClient, data.Name.ValueString(), data.Tags, &response.Diagnostics)
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

func (r *LoadBalancerResource) readLoadBalancer(ctx context.Context, id string, diags diag.Diagnostics) *LoadBalancerModel {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadLoadBalancersByIdResponse, error) {
		return numspotClient.ReadLoadBalancersByIdWithResponse(ctx, r.provider.SpaceID, id)
	}, http.StatusOK, &diags)
	if res == nil {
		return nil
	}

	tf := LoadBalancerFromHttpToTf(ctx, res.JSON200, &diags)
	if diags.HasError() {
		return nil
	}

	return tf
}

func (r *LoadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

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

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
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

		DeleteLoadBalancerTags(ctx, r.provider.SpaceID, numspotClient, state.Name.ValueString(), toDeleteTf, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		toCreateTf, diags = types.ListValueFrom(ctx, tagType, toCreate)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		CreateLoadBalancerTags(ctx, r.provider.SpaceID, numspotClient, state.Name.ValueString(), toCreateTf, &response.Diagnostics)
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
			r.createListeners(ctx, state.Id.ValueString(), toCreate, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}
		}

		if len(toDelete) > 0 {
			r.deleteListeners(ctx, state.Id.ValueString(), toDelete, &response.Diagnostics)
			if response.Diagnostics.HasError() {
				return
			}
		}

		modifs = true
	}

	if !plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck) {
		// HealthCheck:
		if !plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck) {
			res := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
				return numspotClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), numspot.UpdateLoadBalancerJSONRequestBody{
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

func (r *LoadBalancerResource) createListeners(ctx context.Context, loadBalancerId string, listeners []ListenersValue, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}

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
		return numspotClient.CreateLoadBalancerListenersWithResponse(
			ctx,
			r.provider.SpaceID,
			loadBalancerId,
			numspot.CreateLoadBalancerListenersJSONRequestBody{
				Listeners: listenersForCreation,
			},
		)
	}, http.StatusCreated, diags)
}

func (r *LoadBalancerResource) deleteListeners(ctx context.Context, loadBalancerId string, listeners []ListenersValue, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	listenersLoadBalancerPortToDelete := make([]int, 0, len(listeners))
	for _, e := range listeners {
		listenersLoadBalancerPortToDelete = append(listenersLoadBalancerPortToDelete, utils.FromTfInt64ToInt(e.LoadBalancerPort))
	}

	utils.ExecuteRequest(func() (*numspot.DeleteLoadBalancerListenersResponse, error) {
		return numspotClient.DeleteLoadBalancerListenersWithResponse(
			ctx,
			r.provider.SpaceID,
			loadBalancerId,
			numspot.DeleteLoadBalancerListenersJSONRequestBody{
				LoadBalancerPorts: listenersLoadBalancerPortToDelete,
			},
		)
	}, http.StatusNoContent, diags)
}

func (r *LoadBalancerResource) linkBackendMachines(ctx context.Context, lbName string, backendIps, backendVmIds *[]string, diags *diag.Diagnostics) {
	payload := numspot.LinkLoadBalancerBackendMachinesJSONRequestBody{}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	if backendIps != nil && len(*backendIps) > 0 {
		payload.BackendIps = backendIps
	}

	if backendVmIds != nil && len(*backendVmIds) > 0 {
		payload.BackendVmIds = backendVmIds
	}

	utils.ExecuteRequest(func() (*numspot.LinkLoadBalancerBackendMachinesResponse, error) {
		return numspotClient.LinkLoadBalancerBackendMachinesWithResponse(ctx, r.provider.SpaceID, lbName, payload)
	}, http.StatusNoContent, diags)
}

func (r *LoadBalancerResource) AttachHealthCheck(ctx context.Context, lbName string, healthCheck HealthCheckValue, diags *diag.Diagnostics) *numspot.UpdateLoadBalancerResponse {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}
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
		return numspotClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, lbName, payload)
	}, http.StatusOK, diags)

	return updatedLoadBalancer
}

func (r *LoadBalancerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data LoadBalancerModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("Deleting load balancer %s", data.Id))

	// Detach security groups
	emptyList := []string{}
	res := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
		return numspotClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, data.Name.ValueString(), numspot.UpdateLoadBalancerJSONRequestBody{
			SecurityGroups: &emptyList,
		})
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Load Balancer", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}
