package routetable

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	_ resource.Resource                = &RouteTableResource{}
	_ resource.ResourceWithConfigure   = &RouteTableResource{}
	_ resource.ResourceWithImportState = &RouteTableResource{}
)

type RouteTableResource struct {
	provider *client.NumSpotSDK
}

func NewRouteTableResource() resource.Resource {
	return &RouteTableResource{}
}

func (r *RouteTableResource) Configure(ctx context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(*client.NumSpotSDK)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	r.provider = provider
}

func (r *RouteTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func (r *RouteTableResource) Metadata(ctx context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_route_table"
}

func (r *RouteTableResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = RouteTableResourceSchema(ctx)
}

func (r *RouteTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var plan RouteTableModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	payload := deserializeCreateRouteTable(&plan)
	tagsList := tags.TfTagsToApiTags(ctx, plan.Tags)
	routes := deserializeRoutes(ctx, plan.Routes)
	res, err := core.CreateRouteTable(ctx, r.provider, payload, tagsList, routes, plan.SubnetId.ValueStringPointer())
	if err != nil {
		response.Diagnostics.AddError("failed to create route table", err.Error())
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

func (r *RouteTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	res, err := core.ReadRouteTable(ctx, r.provider, data.Id.ValueString())
	if err != nil {
		response.Diagnostics.AddError("failed to read Route Table", err.Error())
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

func (r *RouteTableResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var (
		state, plan RouteTableModel
		routeTable  *numspot.RouteTable
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

	stateTags := tags.TfTagsToApiTags(ctx, state.Tags)
	planTags := tags.TfTagsToApiTags(ctx, plan.Tags)
	if !state.Tags.Equal(plan.Tags) {
		routeTable, err = core.UpdateRouteTableTags(ctx, r.provider, state.Id.ValueString(), stateTags, planTags)
		if err != nil {
			response.Diagnostics.AddError("failed to update route table tags", err.Error())
			return
		}
	}

	stRoutes := deserializeRoutes(ctx, state.Routes)
	plRoutes := deserializeRoutes(ctx, plan.Routes)
	if !state.Routes.Equal(plan.Routes) {
		routeTable, err = core.UpdateRouteTableRoutes(ctx, r.provider, state.Id.ValueString(), stRoutes, plRoutes)
		if err != nil {
			response.Diagnostics.AddError("failed to update route table routes", err.Error())
			return
		}
	}
	if routeTable != nil {
		tf := serializeRouteTable(
			ctx,
			routeTable,
			&response.Diagnostics,
		)
		if response.Diagnostics.HasError() {
			return
		}
		response.Diagnostics.Append(response.State.Set(ctx, &tf)...)

	}
}

func (r *RouteTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data RouteTableModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	links := utils.TfListToGenericList(func(link LinkRouteTablesValue) string {
		return link.Id.ValueString()
	}, ctx, data.LinkRouteTables, &response.Diagnostics)
	if err := core.DeleteRouteTable(ctx, r.provider, data.Id.ValueString(), links); err != nil {
		response.Diagnostics.AddError("failed to delete route table", err.Error())
	}
}

func serializeRouteTableLink(ctx context.Context, link numspot.LinkRouteTable, diags *diag.Diagnostics) LinkRouteTablesValue {
	value, diagnostics := NewLinkRouteTablesValue(
		LinkRouteTablesValue{}.AttributeTypes(ctx),
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

func serializeRoute(ctx context.Context, route numspot.Route, diags *diag.Diagnostics) RoutesValue {
	value, diagnostics := NewRoutesValue(
		RoutesValue{}.AttributeTypes(ctx),
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

func serializeRouteTable(ctx context.Context, http *numspot.RouteTable, diags *diag.Diagnostics) *RouteTableModel {
	var (
		localRoute RoutesValue
		routes     []numspot.Route
		tagsTf     types.List
	)

	if http.Routes == nil {
		return nil
	}
	for _, route := range *http.Routes {
		if route.GatewayId != nil && *route.GatewayId == "local" {
			localRoute = serializeRoute(ctx, route, diags)
			if diags.HasError() {
				return nil
			}
		} else {
			routes = append(routes, route)
		}
	}

	// Routes
	tfRoutes := utils.GenericSetToTfSetValue(ctx, serializeRoute, routes, diags)
	if diags.HasError() {
		return nil
	}

	// Links
	tfLinks := utils.GenericListToTfListValue(ctx, serializeRouteTableLink, *http.LinkRouteTables, diags)
	if diags.HasError() {
		return nil
	}

	// Retrieve Subnet Id:
	var subnetId *string
	for _, assoc := range *http.LinkRouteTables {
		if assoc.SubnetId != nil {
			subnetId = assoc.SubnetId
			break
		}
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	res := RouteTableModel{
		Id:                              types.StringPointerValue(http.Id),
		LinkRouteTables:                 tfLinks,
		VpcId:                           types.StringPointerValue(http.VpcId),
		RoutePropagatingVirtualGateways: types.ListNull(RoutePropagatingVirtualGatewaysValue{}.Type(ctx)),
		Routes:                          tfRoutes,
		SubnetId:                        types.StringPointerValue(subnetId),
		Tags:                            tagsTf,
		LocalRoute:                      localRoute,
	}

	return &res
}

func deserializeCreateRouteTable(tf *RouteTableModel) numspot.CreateRouteTableJSONRequestBody {
	return numspot.CreateRouteTableJSONRequestBody{
		VpcId: tf.VpcId.ValueString(),
	}
}

func deserializeRoutes(ctx context.Context, tfRoutes types.Set) []numspot.Route {
	routes := make([]numspot.Route, len(tfRoutes.Elements()))
	swap := make([]RoutesValue, len(tfRoutes.Elements()))
	tfRoutes.ElementsAs(ctx, &swap, false)
	for i := 0; i < len(swap); i++ {
		obj := numspot.Route{
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
