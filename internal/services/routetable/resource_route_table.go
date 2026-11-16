package routetable

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/routetable/resource_route_table"
	"github.com/numspot/terraform-provider-numspot/internal/services/tags"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &routeTableResource{}
	_ resource.ResourceWithConfigure   = &routeTableResource{}
	_ resource.ResourceWithImportState = &routeTableResource{}
)

// routeTableResource is the Terraform resource implementation for numspot_route_table.
type routeTableResource struct {
	services.BaseResource
}

// NewRouteTableResource returns a new routeTableResource instance.
func NewRouteTableResource() resource.Resource {
	return &routeTableResource{}
}

// Metadata sets the resource type name.
func (r *routeTableResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_route_table"
}

// Schema defines the schema for the resource.
func (r *routeTableResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_route_table.RouteTableResourceSchema(ctx)
}

// Create creates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *routeTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	tagsValue := routetableTags(ctx, plan.Tags)
	numspotRouteTable, err := core.CreateRouteTable(
		ctx,
		r.Provider,
		deserializeCreateRouteTableRequest(&plan),
		routesFromModel(ctx, &plan, &response.Diagnostics),
		subnetIDsFromModel(ctx, &plan, &response.Diagnostics),
		tagsValue,
	)
	if err != nil {
		response.Diagnostics.AddError("unable to create route_table", err.Error())
		return
	}

	state := serializeRouteTable(ctx, numspotRouteTable, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	state.SubnetId = utils.PreserveFromState(plan.SubnetId)

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

// Read reads the current resource state.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *routeTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotRouteTable, err := core.ReadRouteTable(ctx, r.Provider, state.Id.ValueString())
	if err != nil {
		if utils.IsResourceGone(err) {
			tflog.Warn(ctx, "unable to read route_table (not found or access forbidden); removing from state")
			response.State.RemoveResource(ctx)
			return
		}
		response.Diagnostics.AddError("unable to read route_table", err.Error())
		return
	}

	newState := serializeRouteTable(ctx, numspotRouteTable, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.SubnetId = state.SubnetId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Update updates the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *routeTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		err               error
		state, plan       resource_route_table.RouteTableModel
		numspotRouteTable *api.RouteTable
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	routetableID := state.Id.ValueString()
	planTags := routetableTags(ctx, plan.Tags)
	stateTags := routetableTags(ctx, state.Tags)

	if !plan.Routes.Equal(state.Routes) {
		numspotRouteTable, err = core.UpdateRouteTableRoutes(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			routesFromModel(ctx, &state, &response.Diagnostics),
			routesFromModel(ctx, &plan, &response.Diagnostics),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update route_table", err.Error())
			return
		}
	}

	if !plan.SubnetId.Equal(state.SubnetId) || !plan.SubnetIds.Equal(state.SubnetIds) {
		numspotRouteTable, err = core.UpdateRouteTableSubnets(
			ctx,
			r.Provider,
			state.Id.ValueString(),
			subnetIDsFromModel(ctx, &plan, &response.Diagnostics),
		)
		if err != nil {
			response.Diagnostics.AddError("unable to update route_table", err.Error())
			return
		}
	}

	if !plan.Tags.Equal(state.Tags) {
		numspotRouteTable, err = core.UpdateRouteTableTags(ctx, r.Provider, routetableID, stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update route_table tags", err.Error())
			return
		}
	}

	newState := serializeRouteTable(ctx, numspotRouteTable, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	newState.SubnetId = state.SubnetId

	response.Diagnostics.Append(response.State.Set(ctx, &newState)...)
}

// Delete deletes the resource.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (r *routeTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var state resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := core.DeleteRouteTable(ctx, r.Provider, state.Id.ValueString(), linksFromState(ctx, &state, &response.Diagnostics)); err != nil {
		response.Diagnostics.AddError("unable to delete route_table", err.Error())
		return
	}
}

func serializeRouteTable(ctx context.Context, http *api.RouteTable, diags *diag.Diagnostics) resource_route_table.RouteTableModel {
	var tagsTf types.Set
	if http == nil {
		diags.AddError("nil RouteTable response", "the API returned an empty RouteTable")
		return resource_route_table.RouteTableModel{}
	}
	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
	}

	return resource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		LinkRouteTables:                 utils.GenericListToTfListValue(ctx, serializeLinkRouteTable, *http.LinkRouteTables, diags),
		LocalRoute:                      localRouteFromRoutes(ctx, http.Routes, diags),
		RoutePropagatingVirtualGateways: types.ListNull(resource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx)),
		Routes:                          serializeNonLocalRoutes(ctx, http.Routes, diags),
		SubnetIds:                       subnetIDsFromLinks(ctx, http.LinkRouteTables, diags),
		VpcId:                           types.StringPointerValue(http.VpcId),
		Tags:                            tagsTf,
	}
}

func deserializeCreateRouteTableRequest(tf *resource_route_table.RouteTableModel) *api.CreateRouteTableJSONRequestBody {
	return &api.CreateRouteTableJSONRequestBody{
		VpcId: tf.VpcId.ValueString(),
	}
}

func routetableTags(ctx context.Context, tagsSet types.Set) []api.ResourceTag {
	tfTags := make([]resource_route_table.TagsValue, 0, len(tagsSet.Elements()))
	tagsSet.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}

func serializeLinkRouteTable(ctx context.Context, item *api.LinkRouteTable, diags *diag.Diagnostics) resource_route_table.LinkRouteTablesValue {
	value, d := resource_route_table.NewLinkRouteTablesValue(resource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":             types.StringPointerValue(item.Id),
		"main":           types.BoolPointerValue(item.Main),
		"route_table_id": types.StringPointerValue(item.RouteTableId),
		"subnet_id":      types.StringPointerValue(item.SubnetId),
		"vpc_id":         types.StringPointerValue(item.VpcId),
	})
	diags.Append(d...)
	return value
}
