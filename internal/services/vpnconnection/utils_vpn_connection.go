package vpnconnection

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

const VPNConnectionRouteStateDeleted = "deleted"

func VpnConnectionFromTfToHttp(tf *VpnConnectionModel) *numspot.VpnConnection {
	return &numspot.VpnConnection{}
}

func routeFromHTTP(ctx context.Context, elt numspot.RouteLight, diags *diag.Diagnostics) RoutesValue {
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

func phase1OptionsFromHTTP(ctx context.Context, elt *numspot.Phase1Options, diags *diag.Diagnostics) Phase1optionsValue {
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

func phase2OptionsFromHTTP(ctx context.Context, elt *numspot.Phase2Options, diags *diag.Diagnostics) Phase2optionsValue {
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

func vpnOptionsFromHTTP(ctx context.Context, elt *numspot.VpnOptions, diags *diag.Diagnostics) VpnOptionsValue {
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
		phase1Options := phase1OptionsFromHTTP(ctx, elt.Phase1Options, diags)
		if diags.HasError() {
			return VpnOptionsValue{}
		}
		ph1OptsObj, diagnostics := phase1Options.ToObjectValue(ctx)
		diags.Append(diagnostics...)

		vpnOptions.Phase1options = ph1OptsObj
	}

	if elt.Phase2Options != nil {
		phase2Options := phase2OptionsFromHTTP(ctx, elt.Phase2Options, diags)
		ph2OptsObj, diagnostics := phase2Options.ToObjectValue(ctx)
		diags.Append(diagnostics...)

		vpnOptions.Phase2options = ph2OptsObj
	}

	return vpnOptions
}

func VGWTelemetryFromHTTPToTF(ctx context.Context, http numspot.VgwTelemetry, diags *diag.Diagnostics) VgwTelemetriesValue {
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

func VpnConnectionFromHttpToTf(ctx context.Context, http *numspot.VpnConnection, diags *diag.Diagnostics) *VpnConnectionModel {
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
		VpnOptions:                 vpnOptionsFromHTTP(ctx, http.VpnOptions, diags),
		Tags:                       tagsTf,
	}

	if http.Routes != nil {
		// Skip vpn routes with state deleted
		httpRoutes := slices.DeleteFunc(*http.Routes, func(r numspot.RouteLight) bool {
			return *r.State == VPNConnectionRouteStateDeleted
		})
		routes := utils.GenericSetToTfSetValue(ctx, routeFromHTTP, httpRoutes, diags)
		vpnConnectionModel.Routes = routes
	}

	if http.VgwTelemetries != nil {
		vgwTelemetries := utils.GenericListToTfListValue(ctx, VGWTelemetryFromHTTPToTF, *http.VgwTelemetries, diags)
		vpnConnectionModel.VgwTelemetries = vgwTelemetries
	}

	return &vpnConnectionModel
}

func VpnConnectionFromTfToCreateRequest(tf *VpnConnectionModel) numspot.CreateVpnConnectionJSONRequestBody {
	return numspot.CreateVpnConnectionJSONRequestBody{
		ClientGatewayId:  tf.ClientGatewayId.ValueString(),
		ConnectionType:   tf.ConnectionType.ValueString(),
		StaticRoutesOnly: tf.StaticRoutesOnly.ValueBoolPointer(),
		VirtualGatewayId: tf.VirtualGatewayId.ValueString(),
	}
}

func VpnConnectionFromTfToUpdateRequest(ctx context.Context, tf *VpnConnectionModel, diags *diag.Diagnostics) numspot.UpdateVpnConnectionJSONRequestBody {
	var vpnOptions *numspot.VpnOptionsToUpdate

	phase2Options := phase2OptionsToUpdateFromTFToHTTP(ctx, tf.VpnOptions, diags)
	if phase2Options != nil || tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer() != nil {
		vpnOptions = &numspot.VpnOptionsToUpdate{}
	}
	if vpnOptions != nil {
		vpnOptions.Phase2Options = phase2Options
		vpnOptions.TunnelInsideIpRange = tf.VpnOptions.TunnelInsideIpRange.ValueStringPointer()
	}

	return numspot.UpdateVpnConnectionJSONRequestBody{
		VpnOptions: vpnOptions,
	}
}

func phase2OptionsToUpdateFromTFToHTTP(ctx context.Context, vpnOptions VpnOptionsValue, diags *diag.Diagnostics) *numspot.Phase2OptionsToUpdate {
	vpnOptionsValue, diagnostics := NewPhase2optionsValue(vpnOptions.Phase2options.AttributeTypes(ctx), vpnOptions.Phase2options.Attributes())
	diags.Append(diagnostics...)

	if vpnOptionsValue.PreSharedKey.ValueStringPointer() == nil {
		return nil
	}
	return &numspot.Phase2OptionsToUpdate{PreSharedKey: vpnOptionsValue.PreSharedKey.ValueStringPointer()}
}

func VpnConnectionsFromTfToAPIReadParams(ctx context.Context, tf VpnConnectionsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVpnConnectionsParams {
	return numspot.ReadVpnConnectionsParams{
		States:                   utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:                  utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:                utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:                     utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:                      utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		ConnectionTypes:          utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes, diags),
		ClientGatewayIds:         utils.TfStringListToStringPtrList(ctx, tf.ClientGatewayIds, diags),
		RouteDestinationIpRanges: utils.TfStringListToStringPtrList(ctx, tf.RouteDestinationIpRanges, diags),
		StaticRoutesOnly:         utils.FromTfBoolToBoolPtr(tf.StaticRouteOnly),
		BgpAsns:                  utils.TFInt64ListToIntListPointer(ctx, tf.BgpAsns, diags),
		VirtualGatewayIds:        utils.TfStringListToStringPtrList(ctx, tf.VirtualGatewayIds, diags),
	}
}

func VpnConnectionsFromHttpToTfDatasource(ctx context.Context, http *numspot.VpnConnection, diags *diag.Diagnostics) *VpnConnectionModel {
	var (
		routes              = types.SetNull(RoutesValue{}.Type(ctx))
		vgwTelemetriesValue = types.ListNull(VgwTelemetriesValue{}.Type(ctx))
		tagsList            types.List
	)
	if http.Routes != nil {
		routes = utils.GenericSetToTfSetValue(
			ctx,
			routeFromHTTPDatasource,
			*http.Routes,
			diags,
		)
	}
	if http.VgwTelemetries != nil {
		vgwTelemetriesValue = utils.GenericListToTfListValue(
			ctx,
			VGWTelemetryFromHTTPDatasource,
			*http.VgwTelemetries,
			diags,
		)
	}

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	vpnOptions := vpnOptionsFromHTTPDatasource(ctx, http.VpnOptions, diags)

	return &VpnConnectionModel{
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
	}
}

func routeFromHTTPDatasource(ctx context.Context, elt numspot.RouteLight, diags *diag.Diagnostics) RoutesValue {
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

func phase1OptionsFromHTTPDatasource(ctx context.Context, elt *numspot.Phase1Options, diags *diag.Diagnostics) Phase1optionsValue {
	phase1IntegrityAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1IntegrityAlgorithms, diags)
	phase1EncryptionAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase1EncryptionAlgorithms, diags)
	phase1DHGroupNumbers := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase1DhGroupNumbers, diags)
	ikeVersions := utils.FromStringListPointerToTfStringList(ctx, elt.IkeVersions, diags)

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

func phase2OptionsFromHTTPDatasource(ctx context.Context, elt *numspot.Phase2Options, diags *diag.Diagnostics) Phase2optionsValue {
	phase2IntegrityAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2IntegrityAlgorithms, diags)
	phase2EncryptionAlgorithms := utils.FromStringListPointerToTfStringList(ctx, elt.Phase2EncryptionAlgorithms, diags)
	phase2DHGroupNumbers := utils.FromIntListPointerToTfInt64List(ctx, elt.Phase2DhGroupNumbers, diags)

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

func vpnOptionsFromHTTPDatasource(ctx context.Context, elt *numspot.VpnOptions, diags *diag.Diagnostics) VpnOptionsValue {
	if elt == nil {
		return VpnOptionsValue{}
	}
	vpnOptions, diagnostics := NewVpnOptionsValue(
		VpnOptionsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"phase1options":          NewPhase1optionsValueNull(),
			"phase2options":          NewPhase2optionsValueNull(),
			"tunnel_inside_ip_range": types.StringPointerValue(elt.TunnelInsideIpRange),
		})
	diags.Append(diagnostics...)

	if elt.Phase1Options != nil {
		phase1Options := phase1OptionsFromHTTPDatasource(ctx, elt.Phase1Options, diags)
		vpnOptions.Phase1options.Attributes()["phase1options"] = phase1Options
	}

	if elt.Phase2Options != nil {
		phase2Options := phase2OptionsFromHTTPDatasource(ctx, elt.Phase2Options, diags)
		vpnOptions.Phase2options.Attributes()["phase2options"] = phase2Options
	}

	return vpnOptions
}

func VGWTelemetryFromHTTPDatasource(ctx context.Context, http numspot.VgwTelemetry, diags *diag.Diagnostics) VgwTelemetriesValue {
	var lastStateChangeDate string
	if http.LastStateChangeDate != nil {
		lastStateChangeDate = http.LastStateChangeDate.String()
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
