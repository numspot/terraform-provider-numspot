package vpnconnection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/vpnconnection/datasource_vpn_connection"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &vpnconnectionsDataSource{}

// vpnconnectionsDataSource is the Terraform data source implementation for numspot_vpn_connections.
type vpnconnectionsDataSource struct {
	services.BaseDataSource
}

// NewVpnConnectionsDataSource returns a new vpnconnectionsDataSource instance.
func NewVpnConnectionsDataSource() datasource.DataSource {
	return &vpnconnectionsDataSource{}
}

// Metadata sets the data source type name.
func (d *vpnconnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_connections"
}

// Schema defines the schema for the data source.
func (d *vpnconnectionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpn_connection.VpnConnectionDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *vpnconnectionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadVpnConnectionsWithParams(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read vpn_connections", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.VPNConnection) (datasource_vpn_connection.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, valueDiags := datasource_vpn_connection.NewItemsValue(datasource_vpn_connection.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"client_gateway_configuration": types.StringValue(item.ClientGatewayConfiguration),
		"client_gateway_id":            types.StringValue(item.ClientGatewayId.String()),
		"connection_type":              types.StringValue(item.ConnectionType),
		"id":                           types.StringValue(item.Id.String()),
		"routes":                       mappingRoutes(ctx, item, &diags),
		"state":                        types.StringValue(item.State),
		"static_routes_only":           types.BoolValue(item.StaticRoutesOnly),
		"vgw_telemetries":              mappingVgwTelemetries(ctx, item, &diags),
		"virtual_gateway_id":           types.StringValue(item.VirtualGatewayId.String()),
		"vpn_options":                  mappingVpnOption(ctx, item.VpnOptions, &diags),
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingRoute(ctx context.Context, item *api.RouteLight) (datasource_vpn_connection.RoutesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_vpn_connection.NewRoutesValue(datasource_vpn_connection.RoutesValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"destination_ip_range": types.StringValue(item.DestinationIpRange),
		"route_type":           types.StringValue(item.RouteType),
		"state":                types.StringValue(item.State),
	})
	diags.Append(d...)
	return value, diags
}

func mappingRoutes(ctx context.Context, item *api.VPNConnection, diags *diag.Diagnostics) types.List {
	if item.Routes == nil {
		return types.ListNull(datasource_vpn_connection.RoutesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, item.Routes, mappingRoute)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_vpn_connection.RoutesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingVgwTelemetrie(ctx context.Context, item *api.VgwTelemetry) (datasource_vpn_connection.VgwTelemetriesValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_vpn_connection.NewVgwTelemetriesValue(
		datasource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"accepted_route_count":   utils.FromIntToTfInt64(item.AcceptedRouteCount),
			"last_state_change_date": utils.TimeToTfString(item.LastStateChangeDate),
			"outside_ip_address":     types.StringValue(item.OutsideIpAddress),
			"state":                  types.StringValue(item.State),
			"state_description":      types.StringValue(item.StateDescription),
		},
	)
	diags.Append(d...)
	return value, diags
}

func mappingVgwTelemetries(ctx context.Context, item *api.VPNConnection, diags *diag.Diagnostics) types.List {
	if item.VgwTelemetries == nil {
		return types.ListNull(datasource_vpn_connection.VgwTelemetriesValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, item.VgwTelemetries, mappingVgwTelemetrie)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_vpn_connection.VgwTelemetriesValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func mappingPhase1option(ctx context.Context, item *api.Phase1Options, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vpn_connection.NewPhase1optionsValue(
		datasource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"dpd_timeout_action":          types.StringPointerValue(item.DpdTimeoutAction),
			"dpd_timeout_seconds":         utils.FromIntPtrToTfInt64(item.DpdTimeoutSeconds),
			"ike_versions":                utils.FromStringListPointerToTfStringList(ctx, item.IkeVersions, diags),
			"phase1dh_group_numbers":      utils.FromIntListPointerToTfInt64List(ctx, item.Phase1DhGroupNumbers, diags),
			"phase1encryption_algorithms": utils.FromStringListPointerToTfStringList(ctx, item.Phase1EncryptionAlgorithms, diags),
			"phase1integrity_algorithms":  utils.FromStringListPointerToTfStringList(ctx, item.Phase1IntegrityAlgorithms, diags),
			"phase1lifetime_seconds":      utils.FromIntPtrToTfInt64(item.Phase1LifetimeSeconds),
			"replay_window_size":          utils.FromIntPtrToTfInt64(item.ReplayWindowSize),
			"startup_action":              types.StringPointerValue(item.StartupAction),
		},
	)
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingPhase2option(ctx context.Context, item *api.Phase2Options, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vpn_connection.NewPhase2optionsValue(
		datasource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase2dh_group_numbers":      utils.FromIntListPointerToTfInt64List(ctx, item.Phase2DhGroupNumbers, diags),
			"phase2encryption_algorithms": utils.FromStringListPointerToTfStringList(ctx, item.Phase2EncryptionAlgorithms, diags),
			"phase2integrity_algorithms":  utils.FromStringListPointerToTfStringList(ctx, item.Phase2IntegrityAlgorithms, diags),
			"phase2lifetime_seconds":      utils.FromIntPtrToTfInt64(item.Phase2LifetimeSeconds),
			"pre_shared_key":              types.StringPointerValue(item.PreSharedKey),
		},
	)
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func mappingVpnOption(ctx context.Context, item *api.VpnOptions, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(datasource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx))
	}
	value, d := datasource_vpn_connection.NewVpnOptionsValue(datasource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"phase1options":          mappingPhase1option(ctx, item.Phase1Options, diags),
		"phase2options":          mappingPhase2option(ctx, item.Phase2Options, diags),
		"tunnel_inside_ip_range": types.StringPointerValue(item.TunnelInsideIpRange),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}
