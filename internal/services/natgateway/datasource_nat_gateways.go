package natgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/natgateway/datasource_nat_gateway"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &natGatewaysDataSource{}

type natGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

func NewNatGatewaysDataSource() datasource.DataSource {
	return &natGatewaysDataSource{}
}

func (d *natGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	d.provider = services.ConfigureProviderDatasource(request, response)
}

func (d *natGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_gateways"
}

func (d *natGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nat_gateway.NatGatewayDataSourceSchema(ctx)
}

func (d *natGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_nat_gateway.NatGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeNatGatewaysParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotNATGateway, err := core.ReadNATGatewaysWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read nat gateway", err.Error())
		return
	}

	objectItems := utils.SerializeDatasourceItemsWithDiags(ctx, *numSpotNATGateway, &response.Diagnostics, mappingItemsValue)
	if response.Diagnostics.HasError() {
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeNatGatewaysParams(ctx context.Context, tf datasource_nat_gateway.NatGatewayModel, diags *diag.Diagnostics) api.ReadNatGatewayParams {
	return api.ReadNatGatewayParams{
		SubnetIds: utils.ConvertTfListToArrayOfString(ctx, tf.SubnetIds, diags),
		VpcIds:    utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		States:    utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		TagKeys:   utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues: utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:      utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:       utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
	}
}

func mappingItemsValue(ctx context.Context, natGateway api.NatGateway, diags *diag.Diagnostics) (datasource_nat_gateway.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_nat_gateway.ItemsValue{}.Type(ctx))
	publicIpsList := types.List{}

	if natGateway.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *natGateway.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_nat_gateway.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_nat_gateway.ItemsValue{}, serializeDiags
		}
	}

	if natGateway.PublicIps != nil {
		var serializeDiags diag.Diagnostics
		publicIpsList, serializeDiags = mappingNatGatewayPublicIps(ctx, natGateway, diags)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	}

	return datasource_nat_gateway.NewItemsValue(datasource_nat_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":         types.StringValue(utils.ConvertStringPtrToString(natGateway.Id)),
		"public_ips": publicIpsList,
		"state":      types.StringValue(utils.ConvertStringPtrToString(natGateway.State)),
		"subnet_id":  types.StringValue(utils.ConvertStringPtrToString(natGateway.SubnetId)),
		"tags":       tagsList,
		"vpc_id":     types.StringValue(utils.ConvertStringPtrToString(natGateway.VpcId)),
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_nat_gateway.TagsValue, diag.Diagnostics) {
	return datasource_nat_gateway.NewTagsValue(datasource_nat_gateway.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func mappingNatGatewayPublicIps(ctx context.Context, natGateways api.NatGateway, diags *diag.Diagnostics) (types.List, diag.Diagnostics) {
	lt := len(*natGateways.PublicIps)
	elementValue := make([]datasource_nat_gateway.PublicIpsValue, lt)
	for y, publicIp := range *natGateways.PublicIps {
		elementValue[y], *diags = datasource_nat_gateway.NewPublicIpsValue(datasource_nat_gateway.PublicIpsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"public_ip":    types.StringValue(utils.ConvertStringPtrToString(publicIp.PublicIp)),
			"public_ip_id": types.StringValue(utils.ConvertStringPtrToString(publicIp.PublicIpId)),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_nat_gateway.PublicIpsValue).Type(ctx), elementValue)
}
