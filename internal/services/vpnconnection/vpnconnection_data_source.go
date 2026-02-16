package vpnconnection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/vpnconnection/datasource_vpn_connection"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &vpnConnectionsDataSource{}

type vpnConnectionsDataSource struct {
	provider *client.NumSpotSDK
}

func NewVpnConnectionsDataSource() datasource.DataSource {
	return &vpnConnectionsDataSource{}
}

func (d *vpnConnectionsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func (d *vpnConnectionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpn_connections"
}

func (d *vpnConnectionsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpn_connection.VpnConnectionDataSourceSchema(ctx)
}

func (d *vpnConnectionsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotVpnConnections, err := core.ReadVpnConnectionsWithParams(ctx, d.provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read vpn connections", err.Error())
		return
	}

	objectItems := serializeVpnConnections(ctx, numSpotVpnConnections, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func serializeVpnConnections(ctx context.Context, vpnConnections []api.VPNConnection, diags *diag.Diagnostics) datasource_vpn_connection.VpnConnectionModel {
	vpnConnectionsList := types.ListNull(new(datasource_vpn_connection.ItemsValue).Type(ctx))
	var serializeDiags diag.Diagnostics
	ll := len(vpnConnections)

	if ll > 0 {
		itemsValue := make([]datasource_vpn_connection.ItemsValue, ll)

		for i := 0; i < ll; i++ {
			routeList := types.ListNull(new(datasource_vpn_connection.RoutesValue).Type(ctx))
			telemetryList := types.ListNull(new(datasource_vpn_connection.VgwTelemetriesValue).Type(ctx))
			var options types.Object

			if vpnConnections[i].VpnOptions != nil {
				options, serializeDiags = mappingOptions(ctx, *vpnConnections[i].VpnOptions)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
					continue
				}
			}

			if vpnConnections[i].Routes != nil {
				routeList, serializeDiags = mappingRoute(ctx, vpnConnections[i].Routes)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
					continue
				}
			}

			if vpnConnections[i].VgwTelemetries != nil {
				telemetryList, serializeDiags = mappingTelemetry(ctx, vpnConnections[i].VgwTelemetries)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
					continue
				}
			}

			itemsValue[i], serializeDiags = datasource_vpn_connection.NewItemsValue(datasource_vpn_connection.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"client_gateway_configuration": types.StringValue(vpnConnections[i].ClientGatewayConfiguration),
				"client_gateway_id":            types.StringValue(vpnConnections[i].ClientGatewayId.String()),
				"connection_type":              types.StringValue(vpnConnections[i].ConnectionType),
				"id":                           types.StringValue(vpnConnections[i].Id.String()),
				"routes":                       routeList, // list
				"state":                        types.StringValue(vpnConnections[i].State),
				"static_routes_only":           types.BoolValue(vpnConnections[i].StaticRoutesOnly),
				"vgw_telemetries":              telemetryList, // list
				"virtual_gateway_id":           types.StringValue(vpnConnections[i].VirtualGatewayId.String()),
				"vpn_options":                  options,
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}
		vpnConnectionsList, serializeDiags = types.ListValueFrom(ctx, new(datasource_vpn_connection.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return datasource_vpn_connection.VpnConnectionModel{
		Items: vpnConnectionsList,
	}
}

func mappingOptions(ctx context.Context, opt api.VpnOptions) (types.Object, diag.Diagnostics) {
	ret, diags := datasource_vpn_connection.NewVpnOptionsValueNull().ToObjectValue(ctx)
	if diags.HasError() {
		return ret, diags
	}

	phase1 := basetypes.ObjectValue{}
	if opt.Phase1Options != nil {
		phase1, diags = mappingPhase1(ctx, *opt.Phase1Options)
		if diags.HasError() {
			return ret, diags
		}
	}

	phase2 := basetypes.ObjectValue{}
	if opt.Phase2Options != nil {
		phase2, diags = mappingPhase2(ctx, *opt.Phase2Options)
		if diags.HasError() {
			return ret, diags
		}
	}

	options, diags := datasource_vpn_connection.NewVpnOptionsValue(datasource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"phase1options":          phase1,
		"phase2options":          phase2,
		"tunnel_inside_ip_range": types.StringPointerValue(opt.TunnelInsideIpRange),
	})
	if diags.HasError() {
		return ret, diags
	}

	return options.ToObjectValue(ctx)
}

func mappingPhase1(ctx context.Context, phase1 api.Phase1Options) (types.Object, diag.Diagnostics) {
	ret, diags := datasource_vpn_connection.NewPhase1optionsValueNull().ToObjectValue(ctx)
	if diags.HasError() {
		return ret, diags
	}

	ike := utils.FromStringListPointerToTfStringList(ctx, phase1.IkeVersions, &diags)
	if diags.HasError() {
		return ret, diags
	}

	dhGroup := utils.FromIntListPointerToTfInt64List(ctx, phase1.Phase1DhGroupNumbers, &diags)
	if diags.HasError() {
		return ret, diags
	}

	encryption := utils.FromStringListPointerToTfStringList(ctx, phase1.Phase1EncryptionAlgorithms, &diags)
	if diags.HasError() {
		return ret, diags
	}

	integrity := utils.FromStringListPointerToTfStringList(ctx, phase1.Phase1IntegrityAlgorithms, &diags)
	if diags.HasError() {
		return ret, diags
	}

	tmp, diags := datasource_vpn_connection.NewPhase1optionsValue(datasource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"dpd_timeout_action":          types.StringPointerValue(phase1.DpdTimeoutAction),
		"dpd_timeout_seconds":         types.Int64Value(utils.ConvertIntPtrToInt64(phase1.DpdTimeoutSeconds)),
		"ike_versions":                ike,
		"phase1dh_group_numbers":      dhGroup,
		"phase1encryption_algorithms": encryption,
		"phase1integrity_algorithms":  integrity,
		"phase1lifetime_seconds":      types.Int64Value(utils.ConvertIntPtrToInt64(phase1.Phase1LifetimeSeconds)),
		"replay_window_size":          types.Int64Value(utils.ConvertIntPtrToInt64(phase1.ReplayWindowSize)),
		"startup_action":              types.StringPointerValue(phase1.StartupAction),
	})

	ret, diags = tmp.ToObjectValue(ctx)
	if diags.HasError() {
		return ret, diags
	}
	return ret, diags
}

func mappingPhase2(ctx context.Context, phase2 api.Phase2Options) (types.Object, diag.Diagnostics) {
	ret, diags := datasource_vpn_connection.NewPhase2optionsValueNull().ToObjectValue(ctx)
	if diags.HasError() {
		return ret, diags
	}

	dhGroup := utils.FromIntListPointerToTfInt64List(ctx, phase2.Phase2DhGroupNumbers, &diags)
	if diags.HasError() {
		return ret, diags
	}

	encryption := utils.FromStringListPointerToTfStringList(ctx, phase2.Phase2EncryptionAlgorithms, &diags)
	if diags.HasError() {
		return ret, diags
	}

	integrity := utils.FromStringListPointerToTfStringList(ctx, phase2.Phase2IntegrityAlgorithms, &diags)
	if diags.HasError() {
		return ret, diags
	}

	tmp, diags := datasource_vpn_connection.NewPhase2optionsValue(datasource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"phase2dh_group_numbers":      dhGroup,
		"phase2encryption_algorithms": encryption,
		"phase2integrity_algorithms":  integrity,
		"phase2lifetime_seconds":      types.Int64Value(utils.ConvertIntPtrToInt64(phase2.Phase2LifetimeSeconds)),
		"pre_shared_key":              types.StringPointerValue(phase2.PreSharedKey),
	})

	ret, diags = tmp.ToObjectValue(ctx)
	if diags.HasError() {
		return ret, diags
	}

	return ret, diags
}

func mappingRoute(ctx context.Context, routes []api.RouteLight) (types.List, diag.Diagnostics) {
	list := types.ListNull(new(datasource_vpn_connection.RoutesValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len(routes) > 0 {
		ll := len(routes)
		itemsValue := make([]datasource_vpn_connection.RoutesValue, ll)

		for i := 0; i < ll; i++ {
			itemsValue[i], serializeDiags = datasource_vpn_connection.NewRoutesValue(datasource_vpn_connection.RoutesValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"destination_ip_range": types.StringValue(routes[i].DestinationIpRange),
				"route_type":           types.StringValue(routes[i].RouteType),
				"state":                types.StringValue(routes[i].State),
			})
			if serializeDiags.HasError() {
				continue
			}

		}

		list, serializeDiags = types.ListValueFrom(ctx, new(datasource_vpn_connection.RoutesValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			return list, serializeDiags
		}
	}

	return list, serializeDiags
}

func mappingTelemetry(ctx context.Context, telemetry []api.VgwTelemetry) (types.List, diag.Diagnostics) {
	list := types.ListNull(new(datasource_vpn_connection.VgwTelemetriesValue).Type(ctx))
	var serializeDiags diag.Diagnostics

	if len(telemetry) > 0 {
		ll := len(telemetry)
		itemsValue := make([]datasource_vpn_connection.VgwTelemetriesValue, ll)

		for i := 0; i < ll; i++ {
			itemsValue[i], serializeDiags = datasource_vpn_connection.NewVgwTelemetriesValue(datasource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"accepted_route_count":   types.Int64Value(int64(telemetry[i].AcceptedRouteCount)),
				"last_state_change_date": types.StringValue(telemetry[i].LastStateChangeDate.String()),
				"outside_ip_address":     types.StringValue(telemetry[i].OutsideIpAddress),
				"state":                  types.StringValue(telemetry[i].State),
				"state_description":      types.StringValue(telemetry[i].StateDescription),
			})

			if serializeDiags.HasError() {
				return list, serializeDiags
			}
		}
		list, serializeDiags = types.ListValueFrom(ctx, new(datasource_vpn_connection.VgwTelemetriesValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			return list, serializeDiags
		}
	}

	return list, serializeDiags
}
