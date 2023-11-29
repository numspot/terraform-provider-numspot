package route_table

import (
	"context"
	api_client "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api_client"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/common/slice"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns"
)

var (
	_ resource.Resource                = &RouteTableResource{}
	_ resource.ResourceWithConfigure   = &RouteTableResource{}
	_ resource.ResourceWithImportState = &RouteTableResource{}
)

func NewRouteTableResource() resource.Resource {
	return &RouteTableResource{}
}

type RouteTableResource struct {
	client *api_client.ClientWithResponses
}

type RouteTableResourceModel struct {
	ID                              types.String `tfsdk:"id"`
	VPCID                           types.String `tfsdk:"vpc_id"`
	RoutePropagatingVirtualGateways types.List   `tfsdk:"route_propagating_virtual_gateways"`
	LinkRouteTables                 types.List   `tfsdk:"link_route_tables"`
}

type RoutePropagatingVirtualGateway struct {
	ID types.String `tfsdk:"id"`
}

func RoutePropagatingVirtualGatewayType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id": types.StringType,
	}}
}

func RoutePropagatingVirtualGatewaysSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "The route propagating virtual gateways of the NumSpot route table resource.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "The ID of the route propagating virtual gateway.",
					Computed:            true,
				},
			},
		},
		Computed: true,
		Optional: true,
	}
}

type LinkRouteTable struct {
	ID           types.String `tfsdk:"id"`
	Main         types.Bool   `tfsdk:"main"`
	RouteTableID types.String `tfsdk:"route_table_id"`
	SubnetID     types.String `tfsdk:"subnet_id"`
}

func LinkRouteTableType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"id":             types.StringType,
		"main":           types.BoolType,
		"route_table_id": types.StringType,
		"subnet_id":      types.StringType,
	}}
}

func LinkRouteTableSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		MarkdownDescription: "The links of the NumSpot route table resource.",
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"id": schema.StringAttribute{
					MarkdownDescription: "The ID of the link.",
					Computed:            true,
				},
				"main": schema.BoolAttribute{
					MarkdownDescription: "Indicates if it is the main route table or not.",
					Computed:            true,
				},
				"route_table_id": schema.StringAttribute{
					MarkdownDescription: "The ID of the route table.",
					Computed:            true,
				},
				"subnet_id": schema.StringAttribute{
					MarkdownDescription: "The ID of the subnet.",
					Computed:            true,
					Optional:            true,
				},
			},
		},
		Computed: true,
		Optional: true,
	}
}

func (r *RouteTableResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_route_table"
}

func (r *RouteTableResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		MarkdownDescription: "NumSpot route table resource.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot route table resource computed ID.",
				Computed:            true,
			},
			"vpc_id": schema.StringAttribute{
				MarkdownDescription: "The NumSpot route table resource name.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"route_propagating_virtual_gateways": RoutePropagatingVirtualGatewaysSchema(),
			"link_route_tables":                  LinkRouteTableSchema(),
		},
	}
}

func (r *RouteTableResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var routeTablePlan RouteTableResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &routeTablePlan)...)
	if response.Diagnostics.HasError() {
		return
	}

	createRouteTableBody := api_client.CreateRouteTableJSONRequestBody{
		VirtualPrivateCloudId: routeTablePlan.VPCID.ValueString(),
	}

	res, err := r.client.CreateRouteTableWithResponse(ctx, createRouteTableBody)
	if err != nil {
		response.Diagnostics.AddError("Creating Route Table", err.Error())
		return
	}

	numSpotError := conns.HandleErrorBis(http.StatusCreated, res.HTTPResponse.StatusCode, res.Body)
	if numSpotError != nil {
		response.Diagnostics.AddError(numSpotError.Title, numSpotError.Detail)
		return
	}

	routeTable := res.JSON201

	createdRouteTable, diags := mapToRouteTableResourceModel(ctx, routeTable)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &createdRouteTable)...)
}

func (r *RouteTableResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var routeTableState RouteTableResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &routeTableState)...)
	if response.Diagnostics.HasError() {
		return
	}

	routeTables, err := r.client.ListRouteTablesWithResponse(ctx)
	if err != nil {
		response.Diagnostics.AddError("Reading Route Table", err.Error())
		return
	}

	routeTable := slice.FindFirst(
		*routeTables.JSON200.Items,
		func(table api_client.RouteTable) bool { return table.Id == routeTableState.ID.ValueString() },
	)

	if routeTable == nil {
		response.State.RemoveResource(ctx)
		return
	}

	routeTableResourceModel, diags := mapToRouteTableResourceModel(ctx, routeTable)
	if diags.HasError() {
		response.Diagnostics.Append(diags...)
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &routeTableResourceModel)...)
}

func (r *RouteTableResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
	panic("Not supposed to be updated.")
}

func (r *RouteTableResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var routeTableState RouteTableResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &routeTableState)...)
	if response.Diagnostics.HasError() {
		return
	}

	res, err := r.client.DeleteRouteTableWithResponse(ctx, routeTableState.ID.ValueString())
	if err != nil {
		response.Diagnostics.AddError("Deleting Route Table", err.Error())
		return
	}

	numSpotError := conns.HandleErrorBis(http.StatusNoContent, res.HTTPResponse.StatusCode, res.Body)
	if numSpotError != nil {
		response.Diagnostics.AddError(numSpotError.Title, numSpotError.Detail)
		return
	}
}

func (r *RouteTableResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	client := conns.GetClient(request, response)
	if client == nil || response.Diagnostics.HasError() {
		return
	}
	r.client = client
}

func (r *RouteTableResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), request, response)
}

func mapToRouteTableResourceModel(ctx context.Context, routeTable *api_client.RouteTable) (*RouteTableResourceModel, diag.Diagnostics) {
	routePropagatingVirtualGateways, diags := mapToModelRoutePropagatingVirtualGateways(ctx, *routeTable.RoutePropagatingVirtualGateways)
	if diags.HasError() {
		return nil, diags
	}

	linkRouteTables, diags := mapToModelLinkRouteTables(ctx, *routeTable.LinkRouteTables)
	if diags.HasError() {
		return nil, diags
	}

	return &RouteTableResourceModel{
		ID:                              types.StringValue(routeTable.Id),
		VPCID:                           types.StringValue(routeTable.VirtualPrivateCloudId),
		RoutePropagatingVirtualGateways: routePropagatingVirtualGateways,
		LinkRouteTables:                 linkRouteTables,
	}, diags
}

func mapToModelRoutePropagatingVirtualGateways(ctx context.Context, routePropagatingVirtualGateways []api_client.RoutePropagatingVirtualGateway) (basetypes.ListValue, diag.Diagnostics) {
	return conns.MapHttpListToModelList[api_client.RoutePropagatingVirtualGateway, RoutePropagatingVirtualGateway](
		ctx,
		routePropagatingVirtualGateways,
		func(routePropagatingVirtualGateway api_client.RoutePropagatingVirtualGateway) RoutePropagatingVirtualGateway {
			return RoutePropagatingVirtualGateway{
				ID: types.StringValue(routePropagatingVirtualGateway.Id),
			}
		},
		RoutePropagatingVirtualGatewayType(),
	)
}

func mapToModelLinkRouteTables(ctx context.Context, linkRouteTables []api_client.LinkRouteTable) (basetypes.ListValue, diag.Diagnostics) {
	return conns.MapHttpListToModelList[api_client.LinkRouteTable, LinkRouteTable](
		ctx,
		linkRouteTables,
		func(linkRouteTable api_client.LinkRouteTable) LinkRouteTable {
			return LinkRouteTable{
				ID:           types.StringValue(linkRouteTable.Id),
				Main:         types.BoolValue(linkRouteTable.Main),
				RouteTableID: types.StringValue(linkRouteTable.RouteTableId),
				SubnetID:     types.StringPointerValue(linkRouteTable.SubnetId),
			}
		},
		LinkRouteTableType(),
	)
}
