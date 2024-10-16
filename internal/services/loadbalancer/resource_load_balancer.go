package loadbalancer

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"

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
	var plan LoadBalancerModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)
	backendIP := utils.FromTfStringSetToStringList(ctx, plan.BackendIps, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	backendVM := utils.FromTfStringSetToStringList(ctx, plan.BackendVmIds, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	updateNumSpotLoadBalancer := deserializeUpdateNumSpotLoadBalancer(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	createNumSpotLoadBalancer := deserializeCreateNumSpotLoadBalancer(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotLoadBalancer, err := core.CreateLoadBalancer(ctx, r.provider, createNumSpotLoadBalancer, tagsValue, updateNumSpotLoadBalancer.HealthCheck, backendVM, backendIP)
	if err != nil {
		response.Diagnostics.AddError("unable to create load balancer", err.Error())
		return
	}

	//numspotClient, err := r.provider.GetClient(ctx)
	//if err != nil {
	//	response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
	//	return
	//}
	//// Retries create until request response is OK
	//_, err = utils.RetryCreateUntilResourceAvailableWithBody(
	//	ctx,
	//	r.provider.SpaceID,
	//	LoadBalancerFromTfToCreateRequest(ctx, &data, &response.Diagnostics),
	//	numspotClient.CreateLoadBalancerWithResponse)
	//if err != nil {
	//	response.Diagnostics.AddError("Failed to create Load Balancer", err.Error())
	//	return
	//}
	//
	//// Backends
	//if len(data.BackendVmIds.Elements()) > 0 || len(data.BackendIps.Elements()) > 0 {
	//	r.linkBackendMachines(ctx,
	//		data.Name.ValueString(),
	//		utils.TfStringSetToStringPtrSet(ctx, data.BackendIps, &response.Diagnostics),
	//		utils.TfStringSetToStringPtrSet(ctx, data.BackendVmIds, &response.Diagnostics),
	//		&response.Diagnostics,
	//	)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//}
	//
	//// Health Check
	//if !data.HealthCheck.IsUnknown() {
	//	_ = r.AttachHealthCheck(ctx, data.Name.ValueString(), data.HealthCheck, &response.Diagnostics)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//}
	//
	//// Tags
	//// In load balancer case, tags are not handled in the same way as other resources
	//if len(data.Tags.Elements()) > 0 {
	//	CreateLoadBalancerTags(ctx, r.provider.SpaceID, numspotClient, data.Name.ValueString(), data.Tags, &response.Diagnostics)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//}
	//
	//// Update state with updated resource
	//tf := r.readLoadBalancer(ctx, data.Name.ValueString(), response.Diagnostics)
	//if response.Diagnostics.HasError() {
	//	return
	//}

	state := serializeNumSpotLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func (r *LoadBalancerResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state LoadBalancerModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	loadBalancerName := state.Name.ValueString()

	numSpotLoadBalancer, err := core.ReadLoadBalancer(ctx, r.provider, loadBalancerName)
	if err != nil {
		response.Diagnostics.AddError("unable to read load balancer", err.Error())
		return
	}

	newState := serializeNumSpotLoadBalancer(ctx, numSpotLoadBalancer, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

func (r *LoadBalancerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan LoadBalancerModel
		//diags       diag.Diagnostics
		//modifs      = false
	)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	//numspotClient, err := r.provider.GetClient(ctx)
	//if err != nil {
	//	response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
	//	return
	//}
	//
	//if !plan.Tags.IsUnknown() && !plan.Tags.Equal(state.Tags) {
	//	stateTags := make([]tags.TagsValue, 0, len(state.Tags.Elements()))
	//	diags = state.Tags.ElementsAs(ctx, &stateTags, false)
	//	response.Diagnostics.Append(diags...)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	planTags := make([]tags.TagsValue, 0, len(plan.Tags.Elements()))
	//	diags = plan.Tags.ElementsAs(ctx, &planTags, false)
	//	response.Diagnostics.Append(diags...)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	toCreate, toDelete := utils.Diff(stateTags, planTags)
	//	var toDeleteTf, toCreateTf types.List
	//	tagType := tags.TagsValue{}.Type(ctx)
	//
	//	toDeleteTf, diags = types.ListValueFrom(ctx, tagType, toDelete)
	//	response.Diagnostics.Append(diags...)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	DeleteLoadBalancerTags(ctx, r.provider.SpaceID, numspotClient, state.Name.ValueString(), toDeleteTf, &response.Diagnostics)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	toCreateTf, diags = types.ListValueFrom(ctx, tagType, toCreate)
	//	response.Diagnostics.Append(diags...)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	CreateLoadBalancerTags(ctx, r.provider.SpaceID, numspotClient, state.Name.ValueString(), toCreateTf, &response.Diagnostics)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	modifs = true
	//}
	//
	//if !plan.Listeners.IsUnknown() && !plan.Listeners.Equal(state.Listeners) {
	//	stateListeners := make([]ListenersValue, 0, len(state.Listeners.Elements()))
	//	planListeners := make([]ListenersValue, 0, len(plan.Listeners.Elements()))
	//
	//	diags = state.Listeners.ElementsAs(ctx, &stateListeners, false)
	//	response.Diagnostics.Append(diags...)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	diags = plan.Listeners.ElementsAs(ctx, &planListeners, false)
	//	response.Diagnostics.Append(diags...)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	toCreate, toDelete := utils.Diff(stateListeners, planListeners)
	//
	//	if len(toCreate) > 0 {
	//		r.createListeners(ctx, state.Id.ValueString(), toCreate, &response.Diagnostics)
	//		if response.Diagnostics.HasError() {
	//			return
	//		}
	//	}
	//
	//	if len(toDelete) > 0 {
	//		r.deleteListeners(ctx, state.Id.ValueString(), toDelete, &response.Diagnostics)
	//		if response.Diagnostics.HasError() {
	//			return
	//		}
	//	}
	//
	//	modifs = true
	//}
	//
	//if !plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck) {
	//	// HealthCheck:
	//	if !plan.HealthCheck.IsUnknown() && !state.HealthCheck.Equal(plan.HealthCheck) {
	//		res := utils.ExecuteRequest(func() (*numspot.UpdateLoadBalancerResponse, error) {
	//			return numspotClient.UpdateLoadBalancerWithResponse(ctx, r.provider.SpaceID, plan.Name.ValueString(), numspot.UpdateLoadBalancerJSONRequestBody{
	//				HealthCheck: &numspot.HealthCheck{
	//					CheckInterval:      utils.FromTfInt64ToInt(plan.HealthCheck.CheckInterval),
	//					HealthyThreshold:   utils.FromTfInt64ToInt(plan.HealthCheck.HealthyThreshold),
	//					Path:               plan.HealthCheck.Path.ValueStringPointer(),
	//					Port:               utils.FromTfInt64ToInt(plan.HealthCheck.Port),
	//					Protocol:           plan.HealthCheck.Protocol.ValueString(),
	//					Timeout:            utils.FromTfInt64ToInt(plan.HealthCheck.Timeout),
	//					UnhealthyThreshold: utils.FromTfInt64ToInt(plan.HealthCheck.UnhealthyThreshold),
	//				},
	//			})
	//		}, http.StatusOK, &response.Diagnostics)
	//		if res == nil {
	//			return
	//		}
	//	}
	//
	//	modifs = true
	//}
	//
	//if modifs {
	//	tf := r.readLoadBalancer(ctx, state.Id.ValueString(), response.Diagnostics)
	//	if response.Diagnostics.HasError() {
	//		return
	//	}
	//
	//	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
	//}
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
	tflog.Debug(ctx, fmt.Sprintf("Deleting load balancer %s", data.Name))

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

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Name.ValueString(), numspotClient.DeleteLoadBalancerWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete Load Balancer", err.Error())
		return
	}

	response.State.RemoveResource(ctx)
}

func deserializeCreateNumSpotLoadBalancer(ctx context.Context, tf LoadBalancerModel, diags *diag.Diagnostics) numspot.CreateLoadBalancerJSONRequestBody {
	var securityGroupsPtr *[]string
	if !(tf.SecurityGroups.IsNull() || tf.SecurityGroups.IsUnknown()) {
		securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
		securityGroupsPtr = &securityGroups
	}
	subnets := utils.TfStringListToStringList(ctx, tf.Subnets, diags)
	listeners := utils.TfSetToGenericList(func(a ListenersValue) numspot.ListenerForCreation {
		return numspot.ListenerForCreation{
			BackendPort:          utils.FromTfInt64ToInt(a.BackendPort),
			BackendProtocol:      a.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToInt(a.LoadBalancerPort),
			LoadBalancerProtocol: a.LoadBalancerProtocol.ValueString(),
		}
	}, ctx, tf.Listeners, diags)

	return numspot.CreateLoadBalancerJSONRequestBody{
		Listeners:      listeners,
		Name:           tf.Name.ValueString(),
		PublicIp:       tf.PublicIp.ValueStringPointer(),
		SecurityGroups: securityGroupsPtr,
		Subnets:        subnets,
		Type:           tf.Type.ValueStringPointer(),
	}
}
func deserializeUpdateNumSpotLoadBalancer(ctx context.Context, tf LoadBalancerModel, diags *diag.Diagnostics) numspot.UpdateLoadBalancerJSONRequestBody {
	var (
		loadBalancerPort *int                 = nil
		policyNames      *[]string            = nil
		hc               *numspot.HealthCheck = nil
		publicIp         *string              = nil
		securedCookies   *bool                = nil
	)

	if !tf.HealthCheck.IsUnknown() {
		hc = &numspot.HealthCheck{
			CheckInterval:      utils.FromTfInt64ToInt(tf.HealthCheck.CheckInterval),
			HealthyThreshold:   utils.FromTfInt64ToInt(tf.HealthCheck.HealthyThreshold),
			Path:               tf.HealthCheck.Path.ValueStringPointer(),
			Port:               utils.FromTfInt64ToInt(tf.HealthCheck.Port),
			Protocol:           tf.HealthCheck.Protocol.ValueString(),
			Timeout:            utils.FromTfInt64ToInt(tf.HealthCheck.Timeout),
			UnhealthyThreshold: utils.FromTfInt64ToInt(tf.HealthCheck.UnhealthyThreshold),
		}
	}
	if !tf.PublicIp.IsUnknown() {
		publicIp = tf.PublicIp.ValueStringPointer()
	}
	if !tf.SecuredCookies.IsUnknown() {
		securedCookies = tf.SecuredCookies.ValueBoolPointer()
	}
	securityGroups := utils.TfStringListToStringList(ctx, tf.SecurityGroups, diags)
	listeners := utils.TfSetToGenericSet(func(elt ListenersValue) numspot.Listener {
		policyNames := utils.TfStringListToStringList(ctx, elt.PolicyNames, diags)
		return numspot.Listener{
			BackendPort:          utils.FromTfInt64ToIntPtr(elt.BackendPort),
			BackendProtocol:      elt.BackendProtocol.ValueStringPointer(),
			LoadBalancerPort:     utils.FromTfInt64ToIntPtr(elt.LoadBalancerPort),
			LoadBalancerProtocol: elt.BackendProtocol.ValueStringPointer(),
			PolicyNames:          &policyNames,
			ServerCertificateId:  elt.ServerCertificateId.ValueStringPointer(),
		}
	}, ctx, tf.Listeners, diags)

	if len(listeners) == 1 {
		loadBalancerPort = listeners[0].LoadBalancerPort
		policyNames = listeners[0].PolicyNames
	}

	return numspot.UpdateLoadBalancerJSONRequestBody{
		HealthCheck:      hc,
		LoadBalancerPort: loadBalancerPort,
		PolicyNames:      policyNames,
		PublicIp:         publicIp,
		SecuredCookies:   securedCookies,
		SecurityGroups:   &securityGroups,
	}
}

func serializeNumSpotLoadBalancer(ctx context.Context, http *numspot.LoadBalancer, diags *diag.Diagnostics) LoadBalancerModel {
	var tagsTf types.List

	applicationStickyCookiePoliciestypes := utils.GenericListToTfListValue(ctx, ApplicationStickyCookiePoliciesValue{}, applicationStickyCookiePoliciesFromHTTP, *http.ApplicationStickyCookiePolicies, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	listeners := utils.GenericSetToTfSetValue(ctx, ListenersValue{}, listenersFromHTTP, *http.Listeners, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	stickyCookiePolicies := utils.GenericListToTfListValue(ctx, StickyCookiePoliciesValue{}, stickyCookiePoliciesFromHTTP, *http.StickyCookiePolicies, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	backendIps := utils.FromStringListPointerToTfStringSet(ctx, http.BackendIps, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	backendVmIds := utils.FromStringListPointerToTfStringSet(ctx, http.BackendVmIds, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return LoadBalancerModel{}
		}
	}

	healthCheck, diagnostics := NewHealthCheckValue(HealthCheckValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"check_interval":      utils.FromIntToTfInt64(http.HealthCheck.CheckInterval),
			"healthy_threshold":   utils.FromIntToTfInt64(http.HealthCheck.HealthyThreshold),
			"path":                types.StringPointerValue(http.HealthCheck.Path),
			"port":                utils.FromIntToTfInt64(http.HealthCheck.Port),
			"protocol":            types.StringValue(http.HealthCheck.Protocol),
			"timeout":             utils.FromIntToTfInt64(http.HealthCheck.Timeout),
			"unhealthy_threshold": utils.FromIntToTfInt64(http.HealthCheck.UnhealthyThreshold),
		})

	diags.Append(diagnostics...)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	securityGroups := utils.FromStringListPointerToTfStringList(ctx, http.SecurityGroups, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	sourceSecurityGroup := SourceSecurityGroupValue{
		SecurityGroupName: types.StringPointerValue(http.SourceSecurityGroup.SecurityGroupName),
	}

	subnets := utils.FromStringListPointerToTfStringList(ctx, http.Subnets, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	azNames := utils.FromStringListPointerToTfStringList(ctx, http.AvailabilityZoneNames, diags)
	if diags.HasError() {
		return LoadBalancerModel{}
	}

	return LoadBalancerModel{
		ApplicationStickyCookiePolicies: applicationStickyCookiePoliciestypes,
		BackendIps:                      backendIps,
		BackendVmIds:                    backendVmIds,
		DnsName:                         types.StringPointerValue(http.DnsName),
		HealthCheck:                     healthCheck,
		Listeners:                       listeners,
		Name:                            types.StringPointerValue(http.Name),
		VpcId:                           types.StringPointerValue(http.VpcId),
		PublicIp:                        types.StringPointerValue(http.PublicIp),
		SecuredCookies:                  types.BoolPointerValue(http.SecuredCookies),
		SecurityGroups:                  securityGroups,
		SourceSecurityGroup:             sourceSecurityGroup,
		StickyCookiePolicies:            stickyCookiePolicies,
		Subnets:                         subnets,
		AvailabilityZoneNames:           azNames,
		Type:                            types.StringPointerValue(http.Type),
		Tags:                            tagsTf,
	}
}
