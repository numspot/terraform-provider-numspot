package vpnconnection

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
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

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	// Retries create until request response is OK
	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		r.provider.SpaceID,
		vpnConnectionFromTfToCreateRequest(&plan),
		numspotClient.CreateVpnConnectionWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to create VPN Connection", err.Error())
		return
	}

	createdId := *res.JSON201.Id
	if len(plan.Tags.Elements()) > 0 {
		tags.CreateTagsFromTf(ctx, numspotClient, r.provider.SpaceID, &response.Diagnostics, createdId, plan.Tags)
		if response.Diagnostics.HasError() {
			return
		}
	}

	if !utils.IsTfValueNull(plan.Routes) {
		routes := r.routesSetToRoutesSlice(ctx, plan.Routes, &response.Diagnostics)
		r.addRoutes(ctx, createdId, routes, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf := r.updateVPNOptions(ctx, createdId, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf = r.readVPNConnection(ctx, createdId, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *Resource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf := r.readVPNConnection(ctx, data.Id.ValueString(), &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *Resource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var state, plan VpnConnectionModel
	modifications := false

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	if !state.Tags.Equal(plan.Tags) {
		tags.UpdateTags(
			ctx,
			state.Tags,
			plan.Tags,
			&response.Diagnostics,
			numspotClient,
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
		planRoutes := r.routesSetToRoutesSlice(ctx, plan.Routes, &response.Diagnostics)
		stateRoutes := r.routesSetToRoutesSlice(ctx, state.Routes, &response.Diagnostics)

		routesToCreate, routesToDelete := utils.Diff(stateRoutes, planRoutes)
		r.deleteRoutes(ctx, state.Id.ValueString(), routesToDelete, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}

		r.addRoutes(ctx, state.Id.ValueString(), routesToCreate, &response.Diagnostics)
		if response.Diagnostics.HasError() {
			return
		}
	}

	tf := r.readVPNConnection(ctx, state.Id.ValueString(), &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf = r.updateVPNOptions(ctx, state.Id.ValueString(), plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
	if response.Diagnostics.HasError() {
		return
	}

	tf = r.readVPNConnection(ctx, state.Id.ValueString(), &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, tf)...)
}

func (r *Resource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data VpnConnectionModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, r.provider.SpaceID, data.Id.ValueString(), numspotClient.DeleteVpnConnectionWithResponse)
	if err != nil {
		response.Diagnostics.AddError("Failed to delete VPN Connection", err.Error())
		return
	}
}

func (r *Resource) addRoutes(ctx context.Context, vpnID string, tfRoutes []RoutesValue, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	routes := make([]numspot.CreateVpnConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = numspot.CreateVpnConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	for _, route := range routes {
		res, err := numspotClient.CreateVpnConnectionRouteWithResponse(ctx, r.provider.SpaceID, vpnID, route)
		if err != nil {
			diags.AddError("Error while creating VPN Connection Route", err.Error())
			return
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			diags.AddError("Error while parsing VPN Connection Route", err.Error())
			return
		}
	}
}

func (r *Resource) deleteRoutes(ctx context.Context, vpnID string, tfRoutes []RoutesValue, diags *diag.Diagnostics) {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	routes := make([]numspot.DeleteVpnConnectionRoute, len(tfRoutes))
	for i := range tfRoutes {
		routes[i] = numspot.DeleteVpnConnectionRoute{
			DestinationIpRange: tfRoutes[i].DestinationIpRange.ValueString(),
		}
	}

	for _, route := range routes {
		res, err := numspotClient.DeleteVpnConnectionRouteWithResponse(ctx, r.provider.SpaceID, vpnID, route)
		if err != nil {
			diags.AddError("Error while deleting VPN Connection Route", err.Error())
			return
		}
		if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
			diags.AddError("Error while parsing VPN Connection Route", err.Error())
			return
		}
	}
}

func (r *Resource) routesSetToRoutesSlice(ctx context.Context, list types.Set, diags *diag.Diagnostics) []RoutesValue {
	return utils.TfSetToGenericList(func(a RoutesValue) RoutesValue {
		return RoutesValue{
			DestinationIpRange: a.DestinationIpRange,
		}
	}, ctx, list, diags)
}

func (r *Resource) readVPNConnection(
	ctx context.Context,
	id string,
	diags *diag.Diagnostics,
) *VpnConnectionModel {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	// Retries read on resource until state is OK
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		id,
		r.provider.SpaceID,
		[]string{"pending"},
		[]string{"available"},
		numspotClient.ReadVpnConnectionsByIdWithResponse,
	)
	if err != nil {
		diags.AddError("Failed to read VpnConnection", fmt.Sprintf("Error waiting for instance (%s) to be created: %s", id, err))
		return nil
	}

	rr, ok := read.(*numspot.VpnConnection)
	if !ok {
		diags.AddError("Failed to read vpn connection", "object conversion error")
		return nil
	}

	tf := vpnConnectionFromHttpToTf(ctx, rr, diags)

	return tf
}

func (r *Resource) updateVPNOptions(
	ctx context.Context,
	id string,
	plan VpnConnectionModel,
	diags *diag.Diagnostics,
) *VpnConnectionModel {
	numspotClient, err := r.provider.GetClient(ctx)
	if err != nil {
		diags.AddError("Error while initiating numspotClient", err.Error())
		return nil
	}

	res, err := numspotClient.UpdateVpnConnectionWithResponse(ctx, r.provider.SpaceID, id, vpnConnectionFromTfToUpdateRequest(ctx, &plan, diags))
	if err != nil {
		diags.AddError("Error while updating VPN Connection", err.Error())
		return nil
	}
	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		diags.AddError("Error while parsing VPN Connection", err.Error())
		return nil
	}
	tf := vpnConnectionFromHttpToTf(ctx, res.JSON200, diags)

	return tf
}
