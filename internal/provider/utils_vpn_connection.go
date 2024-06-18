package provider

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

const VPNConnectionRouteStateDeleted = "deleted"

func VpnConnectionFromTfToHttp(tf *resource_vpn_connection.VpnConnectionModel) *iaas.VpnConnection {
	return &iaas.VpnConnection{}
}

func routeFromHTTP(ctx context.Context, elt iaas.RouteLight) (resource_vpn_connection.RoutesValue, diag.Diagnostics) {
	return resource_vpn_connection.NewRoutesValue(
		resource_vpn_connection.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"destination_ip_range": types.StringPointerValue(elt.DestinationIpRange),
			"route_type":           types.StringPointerValue(elt.RouteType),
			"state":                types.StringPointerValue(elt.State),
		})
}

func phase1OptionsFromHTTP(ctx context.Context, elt *iaas.Phase1Options) (resource_vpn_connection.Phase1optionsValue, diag.Diagnostics) {
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

func phase2OptionsFromHTTP(ctx context.Context, elt *iaas.Phase2Options) (resource_vpn_connection.Phase2optionsValue, diag.Diagnostics) {
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

func vpnOptionsFromHTTP(ctx context.Context, elt *iaas.VpnOptions) (resource_vpn_connection.VpnOptionsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	if elt == nil {
		return resource_vpn_connection.VpnOptionsValue{}, diag.Diagnostics{}
	}

	phase1OptionsNull, diags := resource_vpn_connection.NewPhase1optionsValueUnknown().ToObjectValue(ctx)
	if diags.HasError() {
		return resource_vpn_connection.VpnOptionsValue{}, diags
	}
	phase2OptionsNull, diags := resource_vpn_connection.NewPhase2optionsValueUnknown().ToObjectValue(ctx)
	if diags.HasError() {
		return resource_vpn_connection.VpnOptionsValue{}, diags
	}
	vpnOptions, diags := resource_vpn_connection.NewVpnOptionsValue(
		resource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase1options":          phase1OptionsNull,
			"phase2options":          phase2OptionsNull,
			"tunnel_inside_ip_range": types.StringPointerValue(elt.TunnelInsideIpRange),
		})

	if elt.Phase1Options != nil {
		phase1Options, diags := phase1OptionsFromHTTP(ctx, elt.Phase1Options)
		if diags.HasError() {
			return resource_vpn_connection.VpnOptionsValue{}, diags
		}
		ph1OptsObj, diags := phase1Options.ToObjectValue(ctx)
		if diags.HasError() {
			return resource_vpn_connection.VpnOptionsValue{}, diags
		}
		vpnOptions.Phase1options = ph1OptsObj
	}

	if elt.Phase2Options != nil {
		phase2Options, diags := phase2OptionsFromHTTP(ctx, elt.Phase2Options)
		if diags.HasError() {
			return resource_vpn_connection.VpnOptionsValue{}, diags
		}
		ph2OptsObj, diags := phase2Options.ToObjectValue(ctx)
		if diags.HasError() {
			return resource_vpn_connection.VpnOptionsValue{}, diags
		}
		vpnOptions.Phase2options = ph2OptsObj
	}

	return vpnOptions, diags
}

func VGWTelemetryFromHTTPToTF(ctx context.Context, http iaas.VgwTelemetry) (resource_vpn_connection.VgwTelemetriesValue, diag.Diagnostics) {
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

func VpnConnectionFromHttpToTf(ctx context.Context, http *iaas.VpnConnection) (*resource_vpn_connection.VpnConnectionModel, diag.Diagnostics) {
	var (
		diags  diag.Diagnostics
		tagsTf types.List
	)

	vpnOptions, diags := vpnOptionsFromHTTP(ctx, http.VpnOptions)
	if diags.HasError() {
		return nil, diags
	}

	if http.Tags != nil {
		tagsTf, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
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
		Tags:                       tagsTf,
	}

	if http.Routes != nil {
		// Skip vpn routes with state deleted
		httpRoutes := slices.DeleteFunc(*http.Routes, func(r iaas.RouteLight) bool {
			return *r.State == VPNConnectionRouteStateDeleted
		})
		routes, diags := utils.GenericSetToTfSetValue(ctx, resource_vpn_connection.RoutesValue{}, routeFromHTTP, httpRoutes)
		if diags.HasError() {
			return nil, diags
		}
		vpnConnectionModel.Routes = routes
	}

	if http.VgwTelemetries != nil {
		vgwTelemetries, diags := utils.GenericListToTfListValue(ctx, resource_vpn_connection.VgwTelemetriesValue{}, VGWTelemetryFromHTTPToTF, *http.VgwTelemetries)
		if diags.HasError() {
			return nil, diags
		}
		vpnConnectionModel.VgwTelemetries = vgwTelemetries
	}

	return &vpnConnectionModel, nil
}

func VpnConnectionFromTfToCreateRequest(tf *resource_vpn_connection.VpnConnectionModel) iaas.CreateVpnConnectionJSONRequestBody {
	return iaas.CreateVpnConnectionJSONRequestBody{
		ClientGatewayId:  tf.ClientGatewayId.ValueString(),
		ConnectionType:   tf.ConnectionType.ValueString(),
		StaticRoutesOnly: tf.StaticRoutesOnly.ValueBoolPointer(),
		VirtualGatewayId: tf.VirtualGatewayId.ValueString(),
	}
}

func VpnConnectionFromTfToUpdateRequest(ctx context.Context, tf *resource_vpn_connection.VpnConnectionModel) iaas.UpdateVpnConnectionJSONRequestBody {
	var vpnOptions *iaas.VpnOptionsToUpdate

	phase2Options := phase2OptionsToUpdateFromTFToHTTP(ctx, tf.VpnOptions)
	if phase2Options != nil || tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer() != nil {
		vpnOptions = &iaas.VpnOptionsToUpdate{}
	}
	if vpnOptions != nil {
		vpnOptions.Phase2Options = phase2Options
		vpnOptions.TunnelInsideIpRange = tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer()
	}

	return iaas.UpdateVpnConnectionJSONRequestBody{
		VpnOptions: vpnOptions,
	}
}

func phase2OptionsToUpdateFromTFToHTTP(ctx context.Context, vpnOptions resource_vpn_connection.VpnOptionsValue) *iaas.Phase2OptionsToUpdate {
	vpnOptionsValue, diags := resource_vpn_connection.NewPhase2optionsValue(vpnOptions.Phase2options.AttributeTypes(ctx), vpnOptions.Phase2options.Attributes())
	if diags.HasError() {
		return nil
	}
	if vpnOptionsValue.PreSharedKey.ValueStringPointer() == nil {
		return nil
	}
	return &iaas.Phase2OptionsToUpdate{PreSharedKey: vpnOptionsValue.PreSharedKey.ValueStringPointer()}
}

func VpnConnectionsFromTfToAPIReadParams(ctx context.Context, tf VpnConnectionsDataSourceModel) iaas.ReadVpnConnectionsParams {
	return iaas.ReadVpnConnectionsParams{
		States:                   utils.TfStringListToStringPtrList(ctx, tf.States),
		TagKeys:                  utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:                utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:                     utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:                      utils.TfStringListToStringPtrList(ctx, tf.Ids),
		ConnectionTypes:          utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes),
		ClientGatewayIds:         utils.TfStringListToStringPtrList(ctx, tf.ClientGatewayIds),
		RouteDestinationIpRanges: utils.TfStringListToStringPtrList(ctx, tf.RouteDestinationIpRanges),
		StaticRoutesOnly:         utils.FromTfBoolToBoolPtr(tf.StaticRouteOnly),
		BgpAsns:                  utils.TFInt64ListToIntListPointer(ctx, tf.BgpAsns),
		VirtualGatewayIds:        utils.TfStringListToStringPtrList(ctx, tf.VirtualGatewayIds),
	}
}

func VpnConnectionsFromHttpToTfDatasource(ctx context.Context, http *iaas.VpnConnection) (*datasource_vpn_connection.VpnConnectionModel, diag.Diagnostics) {
	var (
		routes              = types.ListNull(datasource_vpn_connection.RoutesValue{}.Type(ctx))
		vgwTelemetriesValue = types.ListNull(datasource_vpn_connection.VgwTelemetriesValue{}.Type(ctx))
		diags               diag.Diagnostics
		tagsList            types.List
	)
	if http.Routes != nil {
		routes, diags = utils.GenericListToTfListValue(
			ctx,
			datasource_vpn_connection.RoutesValue{},
			routeFromHTTPDatasource,
			*http.Routes,
		)
		if diags.HasError() {
			return nil, diags
		}
	}
	if http.VgwTelemetries != nil {
		vgwTelemetriesValue, diags = utils.GenericListToTfListValue(
			ctx,
			datasource_vpn_connection.VgwTelemetriesValue{},
			VGWTelemetryFromHTTPDatasource,
			*http.VgwTelemetries,
		)
		if diags.HasError() {
			return nil, diags
		}
	}

	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	vpnOptions, diags := vpnOptionsFromHTTPDatasource(ctx, http.VpnOptions)
	if diags.HasError() {
		return nil, diags
	}

	return &datasource_vpn_connection.VpnConnectionModel{
		ClientGatewayConfiguration: types.StringPointerValue(http.ClientGatewayConfiguration),
		ClientGatewayId:            types.StringPointerValue(http.ClientGatewayId),
		ConnectionType:             types.StringPointerValue(http.ConnectionType),
		Id:                         types.StringPointerValue(http.Id),
		Routes:                     routes,
		State:                      types.StringPointerValue(http.State),
		StaticRoutesOnly:           types.BoolPointerValue(http.StaticRoutesOnly),
		VgwTelemetries:             vgwTelemetriesValue,
		VirtualGatewayId:           types.StringPointerValue(http.VirtualGatewayId),
		VpnOptions:                 vpnOptions,
		Tags:                       tagsList,
	}, nil
}

func routeFromHTTPDatasource(ctx context.Context, elt iaas.RouteLight) (datasource_vpn_connection.RoutesValue, diag.Diagnostics) {
	return datasource_vpn_connection.NewRoutesValue(
		datasource_vpn_connection.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"destination_ip_range": types.StringPointerValue(elt.DestinationIpRange),
			"route_type":           types.StringPointerValue(elt.RouteType),
			"state":                types.StringPointerValue(elt.State),
		})
}

func phase1OptionsFromHTTPDatasource(ctx context.Context, elt *iaas.Phase1Options) (datasource_vpn_connection.Phase1optionsValue, diag.Diagnostics) {
	phase1IntegrityAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1IntegrityAlgorithms)
	if diags.HasError() {
		return datasource_vpn_connection.Phase1optionsValue{}, diags
	}
	phase1EncryptionAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1EncryptionAlgorithms)
	if diags.HasError() {
		return datasource_vpn_connection.Phase1optionsValue{}, diags
	}
	phase1DHGroupNumbers, diags := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase1DhGroupNumbers)
	if diags.HasError() {
		return datasource_vpn_connection.Phase1optionsValue{}, diags
	}
	ikeVersions, diags := utils.FromStringListPointerToTfStringList(ctx, elt.IkeVersions)
	if diags.HasError() {
		return datasource_vpn_connection.Phase1optionsValue{}, diags
	}

	return datasource_vpn_connection.NewPhase1optionsValue(
		datasource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx),
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

func phase2OptionsFromHTTPDatasource(ctx context.Context, elt *iaas.Phase2Options) (datasource_vpn_connection.Phase2optionsValue, diag.Diagnostics) {
	phase2IntegrityAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2IntegrityAlgorithms)
	if diags.HasError() {
		return datasource_vpn_connection.Phase2optionsValue{}, diags
	}
	phase2EncryptionAlgorithms, diags := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2EncryptionAlgorithms)
	if diags.HasError() {
		return datasource_vpn_connection.Phase2optionsValue{}, diags
	}
	phase2DHGroupNumbers, diags := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase2DhGroupNumbers)
	if diags.HasError() {
		return datasource_vpn_connection.Phase2optionsValue{}, diags
	}

	return datasource_vpn_connection.NewPhase2optionsValue(
		datasource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase2dh_group_numbers":      phase2DHGroupNumbers,
			"phase2encryption_algorithms": phase2EncryptionAlgorithms,
			"phase2integrity_algorithms":  phase2IntegrityAlgorithms,
			"phase2lifetime_seconds":      utils.FromIntPtrToTfInt64(elt.Phase2LifetimeSeconds),
			"pre_shared_key":              types.StringPointerValue(elt.PreSharedKey),
		})
}

func vpnOptionsFromHTTPDatasource(ctx context.Context, elt *iaas.VpnOptions) (datasource_vpn_connection.VpnOptionsValue, diag.Diagnostics) {
	if elt == nil {
		return datasource_vpn_connection.VpnOptionsValue{}, diag.Diagnostics{}
	}
	vpnOptions, diags := datasource_vpn_connection.NewVpnOptionsValue(
		datasource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase1options":          datasource_vpn_connection.NewPhase1optionsValueNull(),
			"phase2options":          datasource_vpn_connection.NewPhase2optionsValueNull(),
			"tunnel_inside_ip_range": types.StringPointerValue(elt.TunnelInsideIpRange),
		})

	if elt.Phase1Options != nil {
		phase1Options, diags := phase1OptionsFromHTTPDatasource(ctx, elt.Phase1Options)
		if diags.HasError() {
			return datasource_vpn_connection.VpnOptionsValue{}, diags
		}
		vpnOptions.Phase1options.Attributes()["phase1options"] = phase1Options
	}

	if elt.Phase2Options != nil {
		phase2Options, diags := phase2OptionsFromHTTPDatasource(ctx, elt.Phase2Options)
		if diags.HasError() {
			return datasource_vpn_connection.VpnOptionsValue{}, diags
		}
		vpnOptions.Phase1options.Attributes()["phase2options"] = phase2Options
	}

	return vpnOptions, diags
}

func VGWTelemetryFromHTTPDatasource(ctx context.Context, http iaas.VgwTelemetry) (datasource_vpn_connection.VgwTelemetriesValue, diag.Diagnostics) {
	var lastStateChangeDate string
	if http.LastStateChangeDate != nil {
		lastStateChangeDate = http.LastStateChangeDate.String()
	}
	return datasource_vpn_connection.NewVgwTelemetriesValue(
		datasource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"accepted_route_count":   utils.FromIntPtrToTfInt64(http.AcceptedRouteCount),
			"last_state_change_date": types.StringValue(lastStateChangeDate),
			"outside_ip_address":     types.StringPointerValue(http.OutsideIpAddress),
			"state":                  types.StringPointerValue(http.State),
			"state_description":      types.StringPointerValue(http.StateDescription),
		})
}
