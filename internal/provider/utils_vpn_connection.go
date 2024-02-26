package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func VpnConnectionFromTfToHttp(tf *resource_vpn_connection.VpnConnectionModel) *api.VpnConnectionSchema {
	return &api.VpnConnectionSchema{}
}

func routeFromHTTP(ctx context.Context, elt api.RouteLightSchema) (resource_vpn_connection.RoutesValue, diag.Diagnostics) {
	return resource_vpn_connection.NewRoutesValue(
		resource_vpn_connection.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"destination_ip_range": types.StringPointerValue(elt.DestinationIpRange),
			"route_type":           types.StringPointerValue(elt.RouteType),
			"state":                types.StringPointerValue(elt.State),
		})
}

func phase1OptionsFromHTTP(ctx context.Context, elt *api.Phase1OptionsSchema) (resource_vpn_connection.Phase1optionsValue, diag.Diagnostics) {
	phase1IntegrityAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1IntegrityAlgorithms)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}, diags
	}
	phase1EncryptionAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1EncryptionAlgorithms)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}, diags
	}
	phase1DHGroupNumbers, diags := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase1DhGroupNumbers)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}, diags
	}
	ikeVersions, diags := utils.FromStringListPointerToTfStringList(ctx, elt.IkeVersions)
	if diags.HasError() {
		return resource_vpn_connection.Phase1optionsValue{}, diags
	}

	return resource_vpn_connection.NewPhase1optionsValue(
		resource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx),
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
}

func phase2OptionsFromHTTP(ctx context.Context, elt *api.Phase2OptionsSchema) (resource_vpn_connection.Phase2optionsValue, diag.Diagnostics) {
	phase2IntegrityAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2IntegrityAlgorithms)
	if diags.HasError() {
		return resource_vpn_connection.Phase2optionsValue{}, diags
	}
	phase2EncryptionAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2EncryptionAlgorithms)
	if diags.HasError() {
		return resource_vpn_connection.Phase2optionsValue{}, diags
	}
	phase2DHGroupNumbers, diags := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase2DhGroupNumbers)
	if diags.HasError() {
		return resource_vpn_connection.Phase2optionsValue{}, diags
	}

	return resource_vpn_connection.NewPhase2optionsValue(
		resource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase2dh_group_numbers":      phase2DHGroupNumbers,
			"phase2encryption_algorithms": phase2EncryptionAlgorithms,
			"phase2integrity_algorithms":  phase2IntegrityAlgorithms,
			"phase2lifetime_seconds":      utils.FromIntPtrToTfInt64(elt.Phase2LifetimeSeconds),
			"pre_shared_key":              types.StringPointerValue(elt.PreSharedKey),
		})
}

func vpnOptionsFromHTTP(ctx context.Context, elt *api.VpnOptionsSchema) (resource_vpn_connection.VpnOptionsValue, diag.Diagnostics) {

	if elt == nil {
		return resource_vpn_connection.VpnOptionsValue{}, diag.Diagnostics{}
	}
	vpnOptions, diags := resource_vpn_connection.NewVpnOptionsValue(
		resource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase1options":          resource_vpn_connection.NewPhase1optionsValueNull(),
			"phase2options":          resource_vpn_connection.NewPhase2optionsValueNull(),
			"tunnel_inside_ip_range": types.StringPointerValue(elt.TunnelInsideIpRange),
		})

	if elt.Phase1Options != nil {
		phase1Options, diags := phase1OptionsFromHTTP(ctx, elt.Phase1Options)
		if diags.HasError() {
			return resource_vpn_connection.VpnOptionsValue{}, diags
		}
		vpnOptions.Phase1options.Attributes()["phase1options"] = phase1Options
	}

	if elt.Phase2Options != nil {
		phase2Options, diags := phase2OptionsFromHTTP(ctx, elt.Phase2Options)
		if diags.HasError() {
			return resource_vpn_connection.VpnOptionsValue{}, diags
		}
		vpnOptions.Phase1options.Attributes()["phase2options"] = phase2Options
	}

	return vpnOptions, diags
}

func VGWTelemetryFromHTTPToTF(ctx context.Context, http api.VgwTelemetrySchema) (resource_vpn_connection.VgwTelemetriesValue, diag.Diagnostics) {
	var lastStateChangeDate string
	if http.LastStateChangeDate != nil {
		lastStateChangeDate = http.LastStateChangeDate.String()
	} else {
		lastStateChangeDate = ""
	}
	return resource_vpn_connection.NewVgwTelemetriesValue(
		resource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"accepted_route_count":   utils.FromIntPtrToTfInt64(http.AcceptedRouteCount),
			"last_state_change_date": types.StringValue(lastStateChangeDate),
			"outside_ip_address":     types.StringPointerValue(http.OutsideIpAddress),
			"state":                  types.StringPointerValue(http.State),
			"state_description":      types.StringPointerValue(http.StateDescription),
		})
}

func VpnConnectionFromHttpToTf(ctx context.Context, http *api.VpnConnectionSchema) resource_vpn_connection.VpnConnectionModel {

	vpnOptions, diags := vpnOptionsFromHTTP(ctx, http.VpnOptions)
	if diags.HasError() {
		return resource_vpn_connection.VpnConnectionModel{}
	}

	vpnConnectionModel := resource_vpn_connection.VpnConnectionModel{
		ClientGatewayConfiguration: types.StringPointerValue(http.ClientGatewayConfiguration),
		ClientGatewayId:            types.StringPointerValue(http.ClientGatewayId),
		ConnectionType:             types.StringPointerValue(http.ConnectionType),
		Id:                         types.StringPointerValue(http.Id),
		State:                      types.StringPointerValue(http.State),
		StaticRoutesOnly:           types.BoolPointerValue(http.StaticRoutesOnly),
		VirtualGatewayId:           types.StringPointerValue(http.VirtualGatewayId),
		VpnOptions:                 vpnOptions,
	}

	if http.Routes != nil {
		routes, diags := utils.GenericListToTfListValue(ctx, resource_vpn_connection.RoutesValue{}, routeFromHTTP, *http.Routes)
		if diags.HasError() {
			return resource_vpn_connection.VpnConnectionModel{}
		}
		vpnConnectionModel.Routes = routes
	}

	if http.VgwTelemetries != nil {
		vgwTelemetries, diags := utils.GenericListToTfListValue(ctx, resource_vpn_connection.VgwTelemetriesValue{}, VGWTelemetryFromHTTPToTF, *http.VgwTelemetries)
		if diags.HasError() {
			return resource_vpn_connection.VpnConnectionModel{}
		}
		vpnConnectionModel.VgwTelemetries = vgwTelemetries
	}

	return vpnConnectionModel
}

func VpnConnectionFromTfToCreateRequest(tf *resource_vpn_connection.VpnConnectionModel) api.CreateVpnConnectionJSONRequestBody {
	return api.CreateVpnConnectionJSONRequestBody{}
}
