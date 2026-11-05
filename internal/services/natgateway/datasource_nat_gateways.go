package natgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/natgateway/datasource_nat_gateway"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &natgatewaysDataSource{}

// natgatewaysDataSource is the Terraform data source implementation for numspot_nat_gateways.
type natgatewaysDataSource struct {
	services.BaseDataSource
}

// NewNatGatewaysDataSource returns a new natgatewaysDataSource instance.
func NewNatGatewaysDataSource() datasource.DataSource {
	return &natgatewaysDataSource{}
}

// Metadata sets the data source type name.
func (d *natgatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_gateways"
}

// Schema defines the schema for the data source.
func (d *natgatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nat_gateway.NatGatewayDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *natgatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_nat_gateway.NatGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadNatGateways(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadNatGatewaysWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read nat_gateways", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.NatGateway) (datasource_nat_gateway.ItemsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	tagsList := types.ListNull(datasource_nat_gateway.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_nat_gateway.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_nat_gateway.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_nat_gateway.NewItemsValue(datasource_nat_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":         types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"public_ips": mappingPublicIps(ctx, item, &diags),
		"state":      types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"subnet_id":  types.StringValue(utils.ConvertStringPtrToString(item.SubnetId)),
		"vpc_id":     types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":       tagsList,
	})
	diags.Append(valueDiags...)
	return value, diags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_nat_gateway.TagsValue, diag.Diagnostics) {
	return datasource_nat_gateway.NewTagsValue(datasource_nat_gateway.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingPublicIp(ctx context.Context, item *api.PublicIpLight) (datasource_nat_gateway.PublicIpsValue, diag.Diagnostics) {
	var diags diag.Diagnostics
	value, d := datasource_nat_gateway.NewPublicIpsValue(datasource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"public_ip":    types.StringPointerValue(item.PublicIp),
		"public_ip_id": types.StringPointerValue(item.PublicIpId),
	})
	diags.Append(d...)
	return value, diags
}

func mappingPublicIps(ctx context.Context, item *api.NatGateway, diags *diag.Diagnostics) types.List {
	if item.PublicIps == nil {
		return types.ListNull(datasource_nat_gateway.PublicIpsValue{}.Type(ctx))
	}
	items, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.PublicIps, mappingPublicIp)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
		return types.ListNull(datasource_nat_gateway.PublicIpsValue{}.Type(ctx))
	}
	listValueItems := utils.CreateListValueItems(ctx, items, &serializeDiags)
	if serializeDiags.HasError() {
		diags.Append(serializeDiags...)
	}
	return listValueItems
}

func deserializeReadNatGateways(ctx context.Context, tf *datasource_nat_gateway.NatGatewayModel, diags *diag.Diagnostics) *api.ReadNatGatewayParams {
	return &api.ReadNatGatewayParams{
		States:    utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		SubnetIds: utils.ConvertTfListToArrayOfString(ctx, tf.SubnetIds, diags),
		TagKeys:   utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues: utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:      utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:       utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		VpcIds:    utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
