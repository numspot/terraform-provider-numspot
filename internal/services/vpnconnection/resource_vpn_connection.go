package vpnconnection

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

type Resource struct {
	provider *client.NumSpotSDK
}

func NewVpnConnectionResource() resource.Resource {
	return &Resource{}
}

func (r *Resource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	numSpotClient, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = numSpotClient
}

func (r *Resource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *Resource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpn_connection"
}

func (r *Resource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = VpnConnectionResourceSchema(ctx)
}

func (r *Resource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan VpnConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := tags.TfTagsToApiTags(ctx, plan.Tags)

	routeSlice := routesSetToRoutesSlice(ctx, plan.Routes, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnection, err := core.CreateVpnConnection(ctx, r.provider, deserializeCreateVpnConnection(plan), *deserializeUpdateVpnOptions(ctx, plan), deserializeCreateRoutes(routeSlice), tagsValue)
	if err != nil {
		response.Diagnostics.AddError("unable to create vpn connection", err.Error())
		return
	}

	state := serializeVpnConnection(ctx, vpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnectionID := state.Id.ValueString()

	vpnConnection, err := core.ReadVpnConnection(ctx, r.provider, vpnConnectionID)
	if err != nil {
		return
	}

	newState := serializeVpnConnection(ctx, vpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VpnConnectionModel
	var vpnConnection *numspot.VpnConnection
	var err error

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnectionID := state.Id.ValueString()
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)

	if !state.Tags.Equal(plan.Tags) {
		vpnConnection, err = core.UpdateVpnConnectionTags(
			ctx,
			r.provider,
			stateTags,
			planTags,
			vpnConnectionID,
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update vpn connection tags", err.Error())
			return
		}
	}

	if !state.Routes.Equal(plan.Routes) { // TODO : test is always true, even if state/plan routes are the same => implement a better test
		planRoutes := routesSetToRoutesSlice(ctx, plan.Routes, &response.Diagnostics)
		stateRoutes := routesSetToRoutesSlice(ctx, state.Routes, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		tfRoutesToCreate, tfRoutesToDelete := utils.Diff(stateRoutes, planRoutes)

		vpnConnection, err = core.UpdateVpnConnectionRoutes(ctx, r.provider, deserializeDeleteRoutes(tfRoutesToDelete), deserializeCreateRoutes(tfRoutesToCreate), vpnConnectionID)
		if err != nil {
			response.Diagnostics.AddError("unable to update vpn connection routes", err.Error())
			return
		}
	}

	if !plan.VpnOptions.Equal(state.VpnOptions) { // TODO : test is always true, even if state/plan options are the same => implement a better test
		vpnConnection, err = core.UpdateVpnConnectionAttributes(ctx, r.provider, vpnConnectionID, deserializeUpdateVpnConnection(ctx, plan))
		if err != nil {
			response.Diagnostics.AddError("unable to update vpn connection attributes", err.Error())
			return
		}
	}

	newState := serializeVpnConnection(ctx, vpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	err := core.DeleteVpnConnection(ctx, r.provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to delete vpn connection", err.Error())
		return
	}
}

func routesSetToRoutesSlice(ctx context.Context, list types.Set, diags *diag.Diagnostics) []RoutesValue {
	return utils.TfSetToGenericList(func(a RoutesValue) RoutesValue {
		return RoutesValue{
			DestinationIpRange: a.DestinationIpRange,
		}
	}, ctx, list, diags)
}

func deserializeCreateVpnConnection(tf VpnConnectionModel) numspot.CreateVpnConnectionJSONRequestBody {
	return numspot.CreateVpnConnectionJSONRequestBody{
		ClientGatewayId:  tf.ClientGatewayId.ValueString(),
		ConnectionType:   tf.ConnectionType.ValueString(),
		StaticRoutesOnly: tf.StaticRoutesOnly.ValueBoolPointer(),
		VirtualGatewayId: tf.VirtualGatewayId.ValueString(),
	}
}

func deserializeUpdateVpnConnection(ctx context.Context, tf VpnConnectionModel) numspot.UpdateVpnConnectionJSONRequestBody {
	var vpnOptions *numspot.VpnOptionsToUpdate

	phase2Options := deserializeUpdatePhase2Options(ctx, tf.VpnOptions)
	if phase2Options != nil || tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer() != nil {
		vpnOptions = &numspot.VpnOptionsToUpdate{}
	}
	if vpnOptions != nil {
		vpnOptions.Phase2Options = phase2Options
		vpnOptions.TunnelInsideIpRange = tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer()
	}

	return numspot.UpdateVpnConnectionJSONRequestBody{
		VpnOptions: deserializeUpdateVpnOptions(ctx, tf),
		// ClientGatewayId:  tf.ClientGatewayId.ValueStringPointer(),
		// VirtualGatewayId: tf.VirtualGatewayId.ValueStringPointer(),
	}
}

func deserializeUpdateVpnOptions(ctx context.Context, tf VpnConnectionModel) *numspot.VpnOptionsToUpdate {
	var vpnOptions *numspot.VpnOptionsToUpdate

	phase2Options := deserializeUpdatePhase2Options(ctx, tf.VpnOptions)
	if phase2Options != nil || tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer() != nil {
		vpnOptions = &numspot.VpnOptionsToUpdate{}
		vpnOptions.Phase2Options = phase2Options
		vpnOptions.TunnelInsideIpRange = tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer()
	}

	return vpnOptions
}

func deserializeUpdatePhase2Options(ctx context.Context, vpnOptions VpnOptionsValue) *numspot.Phase2OptionsToUpdate {
	attrtypes := vpnOptions.Phase2options.AttributeTypes(ctx)
	attrVals := vpnOptions.Phase2options.Attributes()

	phase2OptionsTf, diagnostics := NewPhase2optionsValue(attrtypes, attrVals)
	if diagnostics.HasError() {
		return &numspot.Phase2OptionsToUpdate{}
	}

	return &numspot.Phase2OptionsToUpdate{PreSharedKey: phase2OptionsTf.PreSharedKey.ValueStringPointer()}
}

func deserializeCreateRoutes(tfRoutes []RoutesValue) []numspot.CreateVpnConnectionRoute {
	routes := make([]numspot.CreateVpnConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = numspot.CreateVpnConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	return routes
}

func deserializeDeleteRoutes(tfRoutes []RoutesValue) []numspot.DeleteVpnConnectionRoute {
	routes := make([]numspot.DeleteVpnConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = numspot.DeleteVpnConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	return routes
}

func serializeVpnConnection(ctx context.Context, http *numspot.VpnConnection, diags *diag.Diagnostics) *VpnConnectionModel {
	var tagsTf types.List

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	vpnConnectionModel := VpnConnectionModel{
		ClientGatewayConfiguration: types.StringPointerValue(http.ClientGatewayConfiguration),
		ClientGatewayId:            types.StringPointerValue(http.ClientGatewayId),
		ConnectionType:             types.StringPointerValue(http.ConnectionType),
		Id:                         types.StringPointerValue(http.Id),
		State:                      types.StringPointerValue(http.State),
		StaticRoutesOnly:           types.BoolPointerValue(http.StaticRoutesOnly),
		VirtualGatewayId:           types.StringPointerValue(http.VirtualGatewayId),
		VpnOptions:                 serializeVpnOptions(ctx, http.VpnOptions, diags),
		Tags:                       tagsTf,
	}

	if http.Routes != nil {
		// Skip vpn routes with state deleted
		httpRoutes := slices.DeleteFunc(*http.Routes, func(r numspot.RouteLight) bool {
			return *r.State == "deleted"
		})
		routes := utils.GenericSetToTfSetValue(ctx, serializeRoutes, httpRoutes, diags)
		vpnConnectionModel.Routes = routes
	}

	if http.VgwTelemetries != nil {
		vgwTelemetries := utils.GenericListToTfListValue(ctx, serializeVGWTelemetry, *http.VgwTelemetries, diags)
		vpnConnectionModel.VgwTelemetries = vgwTelemetries
	}

	return &vpnConnectionModel
}

func serializeRoutes(ctx context.Context, elt numspot.RouteLight, diags *diag.Diagnostics) RoutesValue {
	value, diagnostics := NewRoutesValue(
		RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"destination_ip_range": types.StringPointerValue(elt.DestinationIpRange),
			"route_type":           types.StringPointerValue(elt.RouteType),
			"state":                types.StringPointerValue(elt.State),
		})
	diags.Append(diagnostics...)
	return value
}

func serializeVpnOptions(ctx context.Context, elt *numspot.VpnOptions, diags *diag.Diagnostics) VpnOptionsValue {
	if elt == nil {
		return VpnOptionsValue{}
	}

	phase1OptionsNull, diagnostics := NewPhase1optionsValueUnknown().ToObjectValue(ctx)
	diags.Append(diagnostics...)
	phase2OptionsNull, diagnostics := NewPhase2optionsValueUnknown().ToObjectValue(ctx)
	diags.Append(diagnostics...)

	vpnOptions, diagnostics := NewVpnOptionsValue(
		VpnOptionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase1options":          phase1OptionsNull,
			"phase2options":          phase2OptionsNull,
			"tunnel_inside_ip_range": types.StringPointerValue(elt.TunnelInsideIpRange),
		})
	diags.Append(diagnostics...)

	if elt.Phase1Options != nil {
		phase1Options := serializePhase1Options(ctx, elt.Phase1Options, diags)
		if diags.HasError() {
			return VpnOptionsValue{}
		}
		ph1OptsObj, diagnostics := phase1Options.ToObjectValue(ctx)
		diags.Append(diagnostics...)

		vpnOptions.Phase1options = ph1OptsObj
	}

	if elt.Phase2Options != nil {
		phase2Options := serializePhase2Options(ctx, elt.Phase2Options, diags)
		ph2OptsObj, diagnostics := phase2Options.ToObjectValue(ctx)
		diags.Append(diagnostics...)

		vpnOptions.Phase2options = ph2OptsObj
	}

	return vpnOptions
}

func serializePhase1Options(ctx context.Context, elt *numspot.Phase1Options, diags *diag.Diagnostics) Phase1optionsValue {
	phase1IntegrityAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1IntegrityAlgorithms, diags)
	if diags.HasError() {
		return Phase1optionsValue{}
	}
	phase1EncryptionAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1EncryptionAlgorithms, diags)
	if diags.HasError() {
		return Phase1optionsValue{}
	}
	phase1DHGroupNumbers := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase1DhGroupNumbers, diags)
	if diags.HasError() {
		return Phase1optionsValue{}
	}
	ikeVersions := utils.FromStringListPointerToTfStringList(ctx, elt.IkeVersions, diags)
	if diags.HasError() {
		return Phase1optionsValue{}
	}

	value, diagnostics := NewPhase1optionsValue(
		Phase1optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"dpd_timeout_action":          types.StringPointerValue(elt.DpdTimeoutAction),
			"dpd_timeout_seconds":         utils.FromIntPtrToTfInt64(elt.DpdTimeoutSeconds),
			"ike_versions":                ikeVersions,
			"phase1dh_group_numbers":      phase1DHGroupNumbers,
			"phase1encryption_algorithms": phase1EncryptionAlgorithms,
			"phase1integrity_algorithms":  phase1IntegrityAlgorithms,
			"phase1lifetime_seconds":      utils.FromIntPtrToTfInt64(elt.Phase1LifetimeSeconds),
			"replay_window_size":          utils.FromIntPtrToTfInt64(elt.ReplayWindowSize),
			"startup_action":              types.StringPointerValue(elt.StartupAction),
		})
	diags.Append(diagnostics...)
	return value
}

func serializePhase2Options(ctx context.Context, elt *numspot.Phase2Options, diags *diag.Diagnostics) Phase2optionsValue {
	phase2IntegrityAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2IntegrityAlgorithms, diags)
	if diags.HasError() {
		return Phase2optionsValue{}
	}
	phase2EncryptionAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2EncryptionAlgorithms, diags)
	if diags.HasError() {
		return Phase2optionsValue{}
	}
	phase2DHGroupNumbers := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase2DhGroupNumbers, diags)
	if diags.HasError() {
		return Phase2optionsValue{}
	}

	value, diagnostics := NewPhase2optionsValue(
		Phase2optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase2dh_group_numbers":      phase2DHGroupNumbers,
			"phase2encryption_algorithms": phase2EncryptionAlgorithms,
			"phase2integrity_algorithms":  phase2IntegrityAlgorithms,
			"phase2lifetime_seconds":      utils.FromIntPtrToTfInt64(elt.Phase2LifetimeSeconds),
			"pre_shared_key":              types.StringPointerValue(elt.PreSharedKey),
		})

	diags.Append(diagnostics...)
	return value
}

func serializeVGWTelemetry(ctx context.Context, http numspot.VgwTelemetry, diags *diag.Diagnostics) VgwTelemetriesValue {
	var lastStateChangeDate string
	if http.LastStateChangeDate != nil {
		lastStateChangeDate = http.LastStateChangeDate.String()
	} else {
		lastStateChangeDate = ""
	}
	value, diagnostics := NewVgwTelemetriesValue(
		VgwTelemetriesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"accepted_route_count":   utils.FromIntPtrToTfInt64(http.AcceptedRouteCount),
			"last_state_change_date": types.StringValue(lastStateChangeDate),
			"outside_ip_address":     types.StringPointerValue(http.OutsideIpAddress),
			"state":                  types.StringPointerValue(http.State),
			"state_description":      types.StringPointerValue(http.StateDescription),
		})
	diags.Append(diagnostics...)
	return value
}
