package routetable

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/routetable/resource_route_table"
	"terraform-provider-numspot/internal/services/tags"
	"terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &routeTableResource{}
	_ resource.ResourceWithConfigure   = &routeTableResource{}
	_ resource.ResourceWithImportState = &routeTableResource{}
)

type routeTableResource struct {
	provider *client.NumSpotSDK
}

func NewRouteTableResource() resource.Resource {
	return &routeTableResource{}
}

func (r *routeTableResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	r.provider = services.ConfigureProviderResource(request, response)
}

func (r *routeTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *routeTableResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_route_table"
}

func (r *routeTableResource) Schema(ctx context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = resource_route_table.RouteTableResourceSchema(ctx)
}

func (r *routeTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan resource_route_table.RouteTableModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	payload := deserializeCreateRouteTable(&plan)
	tagsList := routeTableTags(ctx, plan.Tags)
	routes := deserializeRoutes(ctx, plan.Routes)

	res, err := core.CreateRouteTable(ctx, r.provider, payload, tagsList, routes, plan.SubnetId.ValueStringPointer())
	if err != nil {
		response.Diagnostics.AddError("unable to create route table", err.Error())
		return
	}

	tf := serializeRouteTable(ctx, res, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *routeTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var state resource_route_table.RouteTableModel

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := core.ReadRouteTable(ctx, r.provider, state.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("unable to read route table", err.Error())
		return
	}

	tf := serializeRouteTable(
		ctx,
		res,
		&response.Diagnostics,
	)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *routeTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan resource_route_table.RouteTableModel
		routeTable  *api.RouteTable
		err         error
	)

	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	stateTags := routeTableTags(ctx, state.Tags)
	planTags := routeTableTags(ctx, plan.Tags)
	if !state.Tags.Equal(plan.Tags) {
		routeTable, err = core.UpdateRouteTableTags(ctx, r.provider, state.Id.ValueString(), stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("unable to update route table tags", err.Error())
			return
		}
	}

	stRoutes := deserializeRoutes(ctx, state.Routes)
	plRoutes := deserializeRoutes(ctx, plan.Routes)
	if !state.Routes.Equal(plan.Routes) {
		routeTable, err = core.UpdateRouteTableRoutes(ctx, r.provider, state.Id.ValueString(), stRoutes, plRoutes)
		if err != nil {
			response.Diagnostics.AddError("unable to update route table routes", err.Error())
			return
		}
	}

	tf := serializeRouteTable(ctx, routeTable, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(response.State.Set(ctx, &tf)...)
}

func (r *routeTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resource_route_table.RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	links := utils.TfListToGenericList(func(link resource_route_table.LinkRouteTablesValue) string {
		return link.Id.ValueString()
	}, ctx, data.LinkRouteTables, &response.Diagnostics)
	if err := core.DeleteRouteTable(ctx, r.provider, data.Id.ValueString(), links); err != nil {
		response.Diagnostics.AddError("unable to delete route table", err.Error())
	}
}

func serializeRouteTableLink(ctx context.Context, link api.LinkRouteTable, diags *diag.Diagnostics) resource_route_table.LinkRouteTablesValue {
	value, diagnostics := resource_route_table.NewLinkRouteTablesValue(
		resource_route_table.LinkRouteTablesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":             types.StringPointerValue(link.Id),
			"main":           types.BoolPointerValue(link.Main),
			"route_table_id": types.StringPointerValue(link.RouteTableId),
			"subnet_id":      types.StringPointerValue(link.SubnetId),
			"vpc_id":         types.StringPointerValue(link.VpcId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeRoute(ctx context.Context, route api.Route, diags *diag.Diagnostics) resource_route_table.RoutesValue {
	value, diagnostics := resource_route_table.NewRoutesValue(
		resource_route_table.RoutesValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
			"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_id":                  types.StringPointerValue(route.VmId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeLocalRoute(ctx context.Context, route api.Route, diags *diag.Diagnostics) resource_route_table.LocalRouteValue {
	value, diagnostics := resource_route_table.NewLocalRouteValue(
		resource_route_table.LocalRouteValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"creation_method":        types.StringPointerValue(route.CreationMethod),
			"destination_ip_range":   types.StringPointerValue(route.DestinationIpRange),
			"destination_service_id": types.StringPointerValue(route.DestinationServiceId),
			"gateway_id":             types.StringPointerValue(route.GatewayId),
			"nat_gateway_id":         types.StringPointerValue(route.NatGatewayId),
			"vpc_peering_id":         types.StringPointerValue(route.VpcPeeringId),
			"nic_id":                 types.StringPointerValue(route.NicId),
			"state":                  types.StringPointerValue(route.State),
			"vm_id":                  types.StringPointerValue(route.VmId),
		},
	)
	diags.Append(diagnostics...)
	return value
}

func serializeRouteTable(ctx context.Context, http *api.RouteTable, diags *diag.Diagnostics) *resource_route_table.RouteTableModel {
	var (
		localRoute resource_route_table.LocalRouteValue
		routes     []api.Route
		tagsTf     types.Set
	)

	if http.Routes == nil {
		return nil
	}
	for _, route := range *http.Routes {
		if route.GatewayId != nil && *route.GatewayId == "local" {
			localRoute = serializeLocalRoute(ctx, route, diags)
			if diags.HasError() {
				return nil
			}
		} else {
			routes = append(routes, route)
		}
	}

	tfRoutes := utils.GenericSetToTfSetValue(ctx, serializeRoute, routes, diags)
	if diags.HasError() {
		return nil
	}

	tfLinks := utils.GenericListToTfListValue(ctx, serializeRouteTableLink, *http.LinkRouteTables, diags)
	if diags.HasError() {
		return nil
	}

	var subnetId *string
	for _, assoc := range *http.LinkRouteTables {
		if assoc.SubnetId != nil {
			subnetId = assoc.SubnetId
			break
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericSetToTfSetValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	res := resource_route_table.RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		LinkRouteTables:                 tfLinks,
		VpcId:                           types.StringPointerValue(http.VpcId),
		RoutePropagatingVirtualGateways: types.ListNull(resource_route_table.RoutePropagatingVirtualGatewaysValue{}.Type(ctx)),
		Routes:                          tfRoutes,
		SubnetId:                        types.StringPointerValue(subnetId),
		Tags:                            tagsTf,
		LocalRoute:                      localRoute,
	}

	return &res
}

func deserializeCreateRouteTable(tf *resource_route_table.RouteTableModel) api.CreateRouteTableJSONRequestBody {
	return api.CreateRouteTableJSONRequestBody{
		VpcId: tf.VpcId.ValueString(),
	}
}

func deserializeRoutes(ctx context.Context, tfRoutes types.Set) []api.Route {
	routes := make([]api.Route, len(tfRoutes.Elements()))
	swap := make([]resource_route_table.RoutesValue, len(tfRoutes.Elements()))
	tfRoutes.ElementsAs(ctx, &swap, false)
	for i := 0; i < len(swap); i++ {
		obj := api.Route{
			DestinationIpRange: swap[i].DestinationIpRange.ValueStringPointer(),
			GatewayId:          swap[i].GatewayId.ValueStringPointer(),
			NatGatewayId:       swap[i].NatGatewayId.ValueStringPointer(),
			VpcPeeringId:       swap[i].VpcPeeringId.ValueStringPointer(),
			NicId:              swap[i].NicId.ValueStringPointer(),
			VmId:               swap[i].VmId.ValueStringPointer(),
		}
		routes[i] = obj
	}
	return routes
}

func routeTableTags(ctx context.Context, tags types.Set) []api.ResourceTag {
	tfTags := make([]resource_route_table.TagsValue, 0, len(tags.Elements()))
	tags.ElementsAs(ctx, &tfTags, false)

	apiTags := make([]api.ResourceTag, 0, len(tfTags))
	for _, tfTag := range tfTags {
		apiTags = append(apiTags, api.ResourceTag{
			Key:   tfTag.Key.ValueString(),
			Value: tfTag.Value.ValueString(),
		})
	}

	return apiTags
}
