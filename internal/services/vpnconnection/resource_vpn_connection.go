package vpnconnection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/vpnconnection/resource_vpn_connection"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &vpnConnectionResource{}
	_ resource.ResourceWithConfigure   = &vpnConnectionResource{}
	_ resource.ResourceWithImportState = &vpnConnectionResource{}
)

// vpnConnectionResource is the Terraform resource implementation for numspot_vpn_connection.
type vpnConnectionResource struct {
	services.BaseResource
}

// NewVpnConnectionResource returns a new vpnConnectionResource instance.
func NewVpnConnectionResource() resource.Resource {
	return &vpnConnectionResource{}
}

// Metadata sets the resource type name.
func (r *vpnConnectionResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpn_connection"
}

// Schema defines the schema for the resource.
func (r *vpnConnectionResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpn_connection.VpnConnectionResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpnConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	numspotVpnConnection, err := core.CreateVpnConnection(
		ctx,
		r.Provider,
		deserializeCreateVpnConnectionRequest(&plan),
		routesFromPlan(ctx, &plan, &response.Diagnostics),
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create vpn_connection", err.Error())
		return
	}

	state := serializeVpnConnection(ctx, numspotVpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpnConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotVpnConnection, err := core.ReadVpnConnection(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read vpn_connection", err.Error())
		return
	}

	newState := serializeVpnConnection(ctx, numspotVpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpnConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err                  error
		state, plan          resource_vpn_connection.VpnConnectionModel
		numspotVpnConnection *api.VPNConnection
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !plan.Routes.Equal(state.Routes) {
		numspotVpnConnection, err = core.UpdateVpnConnectionRoutes(
			ctx,
			r.Provider,
			routesToDelete(ctx, &state, &plan, &response.Diagnostics),
			routesToCreate(ctx, &state, &plan, &response.Diagnostics),
			state.Id.ValueString(),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update vpn_connection", err.Error())
			return
		}
	}

	newState := serializeVpnConnection(ctx, numspotVpnConnection, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *vpnConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteVpnConnection(ctx, r.Provider, state.Id.ValueString()); err != nil {
		response.Diagnostics.AddError("unable to delete vpn_connection", err.Error())
		return
	}
}

func serializeVpnConnection(ctx context.Context, http *api.VPNConnection, diags *diag.Diagnostics) resource_vpn_connection.VpnConnectionModel {
	if http == nil {
		diags.AddError("nil VpnConnection response", "the API returned an empty VpnConnection")
		return resource_vpn_connection.VpnConnectionModel{}
	}

	return resource_vpn_connection.VpnConnectionModel{
		ClientGatewayConfiguration: types.StringValue(http.ClientGatewayConfiguration),
		ClientGatewayId:            types.StringValue(http.ClientGatewayId.String()),
		ConnectionType:             types.StringValue(http.ConnectionType),
		Id:                         types.StringValue(http.Id.String()),
		Routes:                     serializeActiveRoutes(ctx, http.Routes, diags),
		State:                      types.StringValue(http.State),
		StaticRoutesOnly:           types.BoolValue(http.StaticRoutesOnly),
		VgwTelemetries:             utils.GenericListToTfListValue(ctx, serializeVgwTelemetrie, http.VgwTelemetries, diags),
		VirtualGatewayId:           types.StringValue(http.VirtualGatewayId.String()),
		VpnOptions:                 serializeVpnOption(ctx, http.VpnOptions, diags),
	}
}

func deserializeCreateVpnConnectionRequest(tf *resource_vpn_connection.VpnConnectionModel) *api.CreateVPNConnectionJSONRequestBody {
	return &api.CreateVPNConnectionJSONRequestBody{
		ClientGatewayId:  utils.FromTfStringToUuid(tf.ClientGatewayId),
		ConnectionType:   tf.ConnectionType.ValueString(),
		StaticRoutesOnly: tf.StaticRoutesOnly.ValueBoolPointer(),
		VirtualGatewayId: utils.FromTfStringToUuid(tf.VirtualGatewayId),
	}
}

func serializeVgwTelemetrie(ctx context.Context, item *api.VgwTelemetry, diags *diag.Diagnostics) resource_vpn_connection.VgwTelemetriesValue {
	value, d := resource_vpn_connection.NewVgwTelemetriesValue(
		resource_vpn_connection.VgwTelemetriesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"accepted_route_count":   utils.FromIntToTfInt64(item.AcceptedRouteCount),
			"last_state_change_date": utils.TimeToTfString(item.LastStateChangeDate),
			"outside_ip_address":     types.StringValue(item.OutsideIpAddress),
			"state":                  types.StringValue(item.State),
			"state_description":      types.StringValue(item.StateDescription),
		},
	)
	diags.Append(d...)
	return value
}

func serializePhase1option(ctx context.Context, item *api.Phase1Options, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx))
	}
	value, d := resource_vpn_connection.NewPhase1optionsValue(resource_vpn_connection.Phase1optionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"dpd_timeout_action":          types.StringPointerValue(item.DpdTimeoutAction),
		"dpd_timeout_seconds":         utils.FromIntPtrToTfInt64(item.DpdTimeoutSeconds),
		"ike_versions":                utils.FromStringListPointerToTfStringList(ctx, item.IkeVersions, diags),
		"phase1dh_group_numbers":      utils.FromIntListPointerToTfInt64List(ctx, item.Phase1DhGroupNumbers, diags),
		"phase1encryption_algorithms": utils.FromStringListPointerToTfStringList(ctx, item.Phase1EncryptionAlgorithms, diags),
		"phase1integrity_algorithms":  utils.FromStringListPointerToTfStringList(ctx, item.Phase1IntegrityAlgorithms, diags),
		"phase1lifetime_seconds":      utils.FromIntPtrToTfInt64(item.Phase1LifetimeSeconds),
		"replay_window_size":          utils.FromIntPtrToTfInt64(item.ReplayWindowSize),
		"startup_action":              types.StringPointerValue(item.StartupAction),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func serializePhase2option(ctx context.Context, item *api.Phase2Options, diags *diag.Diagnostics) types.Object {
	if item == nil {
		return types.ObjectNull(resource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx))
	}
	value, d := resource_vpn_connection.NewPhase2optionsValue(resource_vpn_connection.Phase2optionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"phase2dh_group_numbers":      utils.FromIntListPointerToTfInt64List(ctx, item.Phase2DhGroupNumbers, diags),
		"phase2encryption_algorithms": utils.FromStringListPointerToTfStringList(ctx, item.Phase2EncryptionAlgorithms, diags),
		"phase2integrity_algorithms":  utils.FromStringListPointerToTfStringList(ctx, item.Phase2IntegrityAlgorithms, diags),
		"phase2lifetime_seconds":      utils.FromIntPtrToTfInt64(item.Phase2LifetimeSeconds),
		"pre_shared_key":              types.StringPointerValue(item.PreSharedKey),
	})
	diags.Append(d...)
	obj, d2 := value.ToObjectValue(ctx)
	diags.Append(d2...)
	return obj
}

func serializeVpnOption(ctx context.Context, item *api.VpnOptions, diags *diag.Diagnostics) resource_vpn_connection.VpnOptionsValue {
	if item == nil {
		return resource_vpn_connection.NewVpnOptionsValueNull()
	}
	value, d := resource_vpn_connection.NewVpnOptionsValue(resource_vpn_connection.VpnOptionsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"phase1options":          serializePhase1option(ctx, item.Phase1Options, diags),
		"phase2options":          serializePhase2option(ctx, item.Phase2Options, diags),
		"tunnel_inside_ip_range": types.StringPointerValue(item.TunnelInsideIpRange),
	})
	diags.Append(d...)
	return value
}
