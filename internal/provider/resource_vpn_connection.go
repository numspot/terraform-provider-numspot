package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_vpn_connection"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/retry_utils"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &VpnConnectionResource{}
	_ resource.ResourceWithConfigure   = &VpnConnectionResource{}
	_ resource.ResourceWithImportState = &VpnConnectionResource{}
)

type VpnConnectionResource struct {
	provider Provider
}

func NewVpnConnectionResource() resource.Resource {
	return &VpnConnectionResource{}
}

func (r *VpnConnectionResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	client, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = client
}

func (r *VpnConnectionResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *VpnConnectionResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_vpn_connection"
}

func (r *VpnConnectionResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_vpn_connection.VpnConnectionResourceSchema(ctx)
}

func (r *VpnConnectionResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Retries create until request response is OK
	res, err := retry_utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		VpnConnectionFromTfToCreateRequest(&plan),
		r.provider.IaasClient.CreateVpnConnectionWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPN Connection", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(plan.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, r.provider.IaasClient, r.provider.SpaceID, &response.Diagnostics, createdId, plan.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if !utils.IsTfValueNull(plan.Routes) {
		routes := r.routesSetToRoutesSlice(ctx, plan.Routes)
		diags := r.addRoutes(ctx, createdId, routes)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf, diags := r.updateVPNOptions(ctx, createdId, plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf, diags = r.readVPNConnection(ctx, createdId)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpnConnectionResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf, diags := r.readVPNConnection(ctx, data.Id.ValueString())
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpnConnectionResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan resource_vpn_connection.VpnConnectionModel
	modifications := false

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			r.provider.IaasClient,
			r.provider.SpaceID,
			state.Id.ValueString(),
		)
		if response.Diagnostics.HasError() {
			return
		}
		modifications = true
	}

	if !modifications {
		return
	}

	if !state.Routes.Equal(plan.Routes) {
		planRoutes := r.routesSetToRoutesSlice(ctx, plan.Routes)
		stateRoutes := r.routesSetToRoutesSlice(ctx, state.Routes)

		routesToCreate, routesToDelete := utils.Diff(stateRoutes, planRoutes)
		diags := r.deleteRoutes(ctx, state.Id.ValueString(), routesToDelete)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}

		diags = r.addRoutes(ctx, state.Id.ValueString(), routesToCreate)
		response.Diagnostics.Append(diags...)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf, diags := r.readVPNConnection(ctx, state.Id.ValueString())
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf, diags = r.updateVPNOptions(ctx, state.Id.ValueString(), plan)
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf, diags = r.readVPNConnection(ctx, state.Id.ValueString())
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *VpnConnectionResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_vpn_connection.VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	err := retry_utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), r.provider.IaasClient.DeleteVpnConnectionWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VPN Connection", err.Error())
		return
	}
}

func (r *VpnConnectionResource) addRoutes(ctx context.Context, vpnID string, tfRoutes []resource_vpn_connection.RoutesValue) diag.Diagnostics {
	diags := diag.Diagnostics{}
	routes := make([]iaas.CreateVpnConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = iaas.CreateVpnConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	for _, route := range routes {
		_ = utils.ExecuteRequest(func() (*iaas.CreateVpnConnectionRouteResponse, error) {
			return r.provider.IaasClient.CreateVpnConnectionRouteWithResponse(ctx, r.provider.SpaceID, vpnID, route)
		}, http.StatusOK, &diags)
	}

	return diags
}

func (r *VpnConnectionResource) deleteRoutes(ctx context.Context, vpnID string, tfRoutes []resource_vpn_connection.RoutesValue) diag.Diagnostics {
	diags := diag.Diagnostics{}
	routes := make([]iaas.DeleteVpnConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = iaas.DeleteVpnConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	for _, route := range routes {
		_ = utils.ExecuteRequest(func() (*iaas.DeleteVpnConnectionRouteResponse, error) {
			return r.provider.IaasClient.DeleteVpnConnectionRouteWithResponse(ctx, r.provider.SpaceID, vpnID, route)
		}, http.StatusNoContent, &diags)
	}

	return diags
}

func (r *VpnConnectionResource) routesSetToRoutesSlice(ctx context.Context, list types.Set) []resource_vpn_connection.RoutesValue {
	return utils.TfSetToGenericList(func(a resource_vpn_connection.RoutesValue) resource_vpn_connection.RoutesValue {
		return resource_vpn_connection.RoutesValue{
			DestinationIpRange: a.DestinationIpRange,
		}
	}, ctx, list)
}

func (r *VpnConnectionResource) readVPNConnection(
	ctx context.Context,
	id string,
) (*resource_vpn_connection.VpnConnectionModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	// Retries read on resource until state is OK
	read, err := retry_utils.RetryReadUntilStateValid(
		ctx,
		id,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		r.provider.IaasClient.ReadVpnConnectionsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("Failed to read VpnConnection", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", id, err))
		return nil, diags
	}

	rr, ok := read.(*iaas.VpnConnection)
	if !ok {
		diags.AddError("Failed to read vpn connection", "object conversion error")
		return nil, diags
	}

	tf, d := VpnConnectionFromHttpToTf(ctx, rr)
	if d.HasError() {
		diags.Append(d...)
		return nil, diags
	}

	return tf, diags
}

func (r *VpnConnectionResource) updateVPNOptions(
	ctx context.Context,
	id string,
	plan resource_vpn_connection.VpnConnectionModel,
) (*resource_vpn_connection.VpnConnectionModel, diag.Diagnostics) {
	diags := diag.Diagnostics{}
	res := utils.ExecuteRequest(func() (*iaas.UpdateVpnConnectionResponse, error) {
		return r.provider.IaasClient.UpdateVpnConnectionWithResponse(
			ctx,
			r.provider.SpaceID,
			id,
			VpnConnectionFromTfToUpdateRequest(ctx, &plan))
	}, http.StatusOK, &diags)
	if res == nil {
		return nil, diags
	}

	tf, d := VpnConnectionFromHttpToTf(ctx, res.JSON200)
	if d.HasError() {
		diags.Append(d...)
		return nil, diags
	}

	return tf, diags
}
