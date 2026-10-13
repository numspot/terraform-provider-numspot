package virtualgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/virtualgateway/datasource_virtual_gateway"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &virtualgatewaysDataSource{}

// virtualgatewaysDataSource is the Terraform data source implementation for numspot_virtual_gateways.
type virtualgatewaysDataSource struct {
	services.BaseDataSource
}

// NewVirtualGatewaysDataSource returns a new virtualgatewaysDataSource instance.
func NewVirtualGatewaysDataSource() datasource.DataSource {
	return &virtualgatewaysDataSource{}
}

// Metadata sets the data source type name.
func (d *virtualgatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_gateways"
}

// Schema defines the schema for the data source.
func (d *virtualgatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_virtual_gateway.VirtualGatewayDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *virtualgatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_virtual_gateway.VirtualGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	result, err := core.ReadVirtualGatewaysWithParams(ctx, d.Provider)
	if err != nil {
		response.Diagnostics.AddError("unable to read virtual_gateways", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.VirtualGateway) (datasource_virtual_gateway.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, valueDiags := datasource_virtual_gateway.NewItemsValue(datasource_virtual_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"connection_type":              types.StringValue(item.ConnectionType),
		"id":                           types.StringValue(item.Id.String()),
		"state":                        types.StringValue(item.State),
		"vpc_to_virtual_gateway_links": mappingVpcToVirtualGatewayLinks(ctx, item, &diags),
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingVpcToVirtualGatewayLink(
	ctx context.Context,
	item *api.VpcToVirtualGatewayLink,
) (datasource_virtual_gateway.VpcToVirtualGatewayLinksValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_virtual_gateway.NewVpcToVirtualGatewayLinksValue(
		datasource_virtual_gateway.VpcToVirtualGatewayLinksValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"state":  types.StringValue(item.State),
			"vpc_id": types.StringValue(item.VpcId),
		},
	)
	diags.Append(d...)
	return value, diags
}

func mappingVpcToVirtualGatewayLinks(ctx context.Context, item *api.VirtualGateway, diags *diag.Diagnostics) types.List {
	if item.VpcToVirtualGatewayLinks == nil {
		return types.ListNull(datasource_virtual_gateway.VpcToVirtualGatewayLinksValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, item.VpcToVirtualGatewayLinks, mappingVpcToVirtualGatewayLink)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_virtual_gateway.VpcToVirtualGatewayLinksValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}
