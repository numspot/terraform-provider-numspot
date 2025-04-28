package vpnconnection

import (
	"context"
	"slices"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/vpnconnection/resource_vpn_connection"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &vpnConnectionResource{}
	_ resource.ResourceWithConfigure   = &vpnConnectionResource{}
	_ resource.ResourceWithImportState = &vpnConnectionResource{}
)

type vpnConnectionResource struct {
	provider *client.NumSpotSDK
}

func NewVpnConnectionResource() resource.Resource {
	return &vpnConnectionResource{}
}

func (r *vpnConnectionResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *vpnConnectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *vpnConnectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpn_connection"
}

func (r *vpnConnectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpn_connection.VpnConnectionResourceSchema(ctx)
}

func (r *vpnConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	routeSlice := routesSetToRoutesSlice(ctx, plan.Routes, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnection, err := core.CreateVpnConnection(ctx, r.provider, deserializeCreateVpnConnection(plan), deserializeCreateRoutes(routeSlice))
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

func (r *vpnConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnectionID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}

	vpnConnection, err := core.ReadVpnConnection(ctx, r.provider, vpnConnectionID)
	if err != nil {
		response.Diagnostics.AddError("unable to read Vpn connection", err.Error())
		return
	}

	newState := serializeVpnConnection(ctx, vpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, newState)...)
}

func (r *vpnConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_vpn_connection.VpnConnectionModel
	var vpnConnection *api.VPNConnection
	var err error

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnectionID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}

	if !state.Routes.Equal(plan.Routes) {
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

		newState := serializeVpnConnection(ctx, vpnConnection, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		response.Diagnostics.Append(response.State.Set(ctx, newState)...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func (r *vpnConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	vpnConnectionID, err := uuid.Parse(state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to parse id from state", err.Error())
		return
	}

	err = core.DeleteVpnConnection(ctx, r.provider, vpnConnectionID)
	if err != nil {
		response.Diagnostics.AddError("unable to delete vpn connection", err.Error())
		return
	}
}

func routesSetToRoutesSlice(ctx context.Context, list types.Set, diags *diag.Diagnostics) []resource_vpn_connection.RoutesValue {
	return utils.TfSetToGenericList(func(a resource_vpn_connection.RoutesValue) resource_vpn_connection.RoutesValue {
		return resource_vpn_connection.RoutesValue{
			DestinationIpRange: a.DestinationIpRange,
		}
	}, ctx, list, diags)
}

func deserializeCreateVpnConnection(tf resource_vpn_connection.VpnConnectionModel) api.CreateVPNConnectionJSONRequestBody {
	return api.CreateVPNConnectionJSONRequestBody{
		ClientGatewayId:  uuid.MustParse(tf.ClientGatewayId.ValueString()),
		ConnectionType:   tf.ConnectionType.ValueString(),
		StaticRoutesOnly: tf.StaticRoutesOnly.ValueBoolPointer(),
		VirtualGatewayId: uuid.MustParse(tf.VirtualGatewayId.ValueString()),
	}
}

func deserializeCreateRoutes(tfRoutes []resource_vpn_connection.RoutesValue) []api.CreateVPNConnectionRoute {
	routes := make([]api.CreateVPNConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = api.CreateVPNConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	return routes
}

func deserializeDeleteRoutes(tfRoutes []resource_vpn_connection.RoutesValue) []api.DeleteVPNConnectionRoute {
	routes := make([]api.DeleteVPNConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = api.DeleteVPNConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	return routes
}

func serializeVpnConnection(ctx context.Context, http *api.VPNConnection, diags *diag.Diagnostics) *resource_vpn_connection.VpnConnectionModel {
	vpnConnectionModel := resource_vpn_connection.VpnConnectionModel{
		ClientGatewayConfiguration: types.StringValue(http.ClientGatewayConfiguration),
		ClientGatewayId:            types.StringValue(http.ClientGatewayId.String()),
		ConnectionType:             types.StringValue(http.ConnectionType),
		Id:                         types.StringValue(http.Id.String()),
		State:                      types.StringValue(http.State),
		StaticRoutesOnly:           types.BoolValue(http.StaticRoutesOnly),
		VirtualGatewayId:           types.StringValue(http.VirtualGatewayId.String()),
		VpnOptions:                 serializeVpnOptions(ctx, &http.VpnOptions, diags),
	}

	if http.Routes != nil {
		// Skip vpn routes with state deleted
		httpRoutes := slices.DeleteFunc(http.Routes, func(r api.RouteLight) bool {
			return r.State == "deleted"
		})
		routes := utils.GenericSetToTfSetValue(ctx, serializeRoutes, httpRoutes, diags)
		vpnConnectionModel.Routes = routes
	}

	if http.VgwTelemetries != nil {
		vgwTelemetries := utils.GenericListToTfListValue(ctx, serializeVGWTelemetry, http.VgwTelemetries, diags)
		vpnConnectionModel.VgwTelemetries = vgwTelemetries
	}

	return &vpnConnectionModel
}

func serializeRoutes(ctx context.Context, elt api.RouteLight, diags *diag.Diagnostics) resource_vpn_connection.RoutesValue {
	value, diagnostics := resource_vpn_connection.NewRoutesValue(
		resource_vpn_connection.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"destination_ip_range": types.StringValue(elt.DestinationIpRange),
			"route_type":           types.StringValue(elt.RouteType),
			"state":                types.StringValue(elt.State),
		})
	diags.Append(diagnostics...)
	return value
}

func serializeVpnOptions(ctx context.Context, elt *api.VpnOptions, diags *diag.Diagnostics) resource_vpn_connection.VpnOptionsValue {
	if elt == nil {
		return resource_vpn_connection.VpnOptionsValue{}
	}

	phase1OptionsNull, diagnostics := resource_vpn_connection.NewPhase1optionsValueUnknown().ToObjectValue(ctx)
	diags.Append(diagnostics...)
	phase2OptionsNull, diagnostics := resource_vpn_connection.NewPhase2optionsValueUnknown().ToObjectValue(ctx)
	diags.Append(diagnostics...)

	vpnOptions, diagnostics := resource_vpn_connection.NewVpnOptionsValue(
		resource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase1options":          phase1OptionsNull,
			"phase2options":          phase2OptionsNull,
			"tunnel_inside_ip_range": types.StringValue(elt.TunnelInsideIpRange),
		})
	diags.Append(diagnostics...)

	phase1Options := serializePhase1Options(ctx, &elt.Phase1Options, diags)
	if diags.HasError() {
		return resource_vpn_connection.VpnOptionsValue{}
	}
	ph1OptsObj, diagnostics := phase1Options.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	vpnOptions.Phase1options = ph1OptsObj

	phase2Options := serializePhase2Options(ctx, &elt.Phase2Options, diags)
	ph2OptsObj, diagnostics := phase2Options.ToObjectValue(ctx)
	diags.Append(diagnostics...)

	vpnOptions.Phase2options = ph2OptsObj

	return vpnOptions
}

func serializePhase1Options(ctx context.Context, elt *api.Phase1Options, diags *diag.Diagnostics) resource_vpn_connection.Phase1optionsValue {
	phase1IntegrityAlgorithms := utils.FromStringListToTfStringList(ctx, elt.Phase1IntegrityAlgorithms, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}
	}
	phase1EncryptionAlgorithms := utils.FromStringListToTfStringList(ctx, elt.Phase1EncryptionAlgorithms, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}
	}
	phase1DHGroupNumbers := utils.FromIntListToTfInt64List(ctx, elt.Phase1DhGroupNumbers, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}
	}
	ikeVersions := utils.FromStringListToTfStringList(ctx, elt.IkeVersions, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}
	}

	value, diagnostics := resource_vpn_connection.NewPhase1optionsValue(
		resource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"dpd_timeout_action":          types.StringValue(elt.DpdTimeoutAction),
			"dpd_timeout_seconds":         types.Int64Value(int64(elt.DpdTimeoutSeconds)),
			"ike_versions":                ikeVersions,
			"phase1dh_group_numbers":      phase1DHGroupNumbers,
			"phase1encryption_algorithms": phase1EncryptionAlgorithms,
			"phase1integrity_algorithms":  phase1IntegrityAlgorithms,
			"phase1lifetime_seconds":      types.Int64Value(int64(elt.Phase1LifetimeSeconds)),
			"replay_window_size":          types.Int64Value(int64(elt.ReplayWindowSize)),
			"startup_action":              types.StringValue(elt.StartupAction),
		})
	diags.Append(diagnostics...)
	return value
}

func serializePhase2Options(ctx context.Context, elt *api.Phase2Options, diags *diag.Diagnostics) resource_vpn_connection.Phase2optionsValue {
	phase2IntegrityAlgorithms := utils.FromStringListToTfStringList(ctx, elt.Phase2IntegrityAlgorithms, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase2optionsValue{}
	}
	phase2EncryptionAlgorithms := utils.FromStringListToTfStringList(ctx, elt.Phase2EncryptionAlgorithms, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase2optionsValue{}
	}
	phase2DHGroupNumbers := utils.FromIntListToTfInt64List(ctx, elt.Phase2DhGroupNumbers, diags)
	if diags.HasError() {
		return resource_vpn_connection.Phase2optionsValue{}
	}

	value, diagnostics := resource_vpn_connection.NewPhase2optionsValue(
		resource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase2dh_group_numbers":      phase2DHGroupNumbers,
			"phase2encryption_algorithms": phase2EncryptionAlgorithms,
			"phase2integrity_algorithms":  phase2IntegrityAlgorithms,
			"phase2lifetime_seconds":      types.Int64Value(int64(elt.Phase2LifetimeSeconds)),
			"pre_shared_key":              types.StringValue(elt.PreSharedKey),
		})

	diags.Append(diagnostics...)
	return value
}

func serializeVGWTelemetry(ctx context.Context, http api.VgwTelemetry, diags *diag.Diagnostics) resource_vpn_connection.VgwTelemetriesValue {
	var lastStateChangeDate string
	if !http.LastStateChangeDate.IsZero() {
		lastStateChangeDate = http.LastStateChangeDate.String()
	} else {
		lastStateChangeDate = ""
	}
	value, diagnostics := resource_vpn_connection.NewVgwTelemetriesValue(
		resource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"accepted_route_count":   types.Int64Value(int64(http.AcceptedRouteCount)),
			"last_state_change_date": types.StringValue(lastStateChangeDate),
			"outside_ip_address":     types.StringValue(http.OutsideIpAddress),
			"state":                  types.StringValue(http.State),
			"state_description":      types.StringValue(http.StateDescription),
		})
	diags.Append(diagnostics...)
	return value
}
