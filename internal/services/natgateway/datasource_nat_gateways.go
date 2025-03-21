package natgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/natgateway/datasource_nat_gateway"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &natGatewaysDataSource{}
)

func (d *natGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

	d.provider = provider
}

func NewNatGatewaysDataSource() datasource.DataSource {
	return &natGatewaysDataSource{}
}

type natGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *natGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_gateways"
}

// Schema defines the schema for the data source.
func (d *natGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_nat_gateway.NatGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
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

	objectItems := serializeNATGateways(ctx, numSpotNATGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func serializeNATGateways(ctx context.Context, natGateways *[]api.NatGateway, diags *diag.Diagnostics) datasource_nat_gateway.NatGatewayModel {
	var natGatewaysList types.List
	var serializeDiags diag.Diagnostics
	tagsList := types.List{}
	publicIpsList := types.List{}

	if len(*natGateways) != 0 {
		ll := len(*natGateways)
		itemsValue := make([]datasource_nat_gateway.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			if (*natGateways)[i].Tags != nil {
				tagsList, serializeDiags = mappingNatGatewayTags(ctx, natGateways, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			if (*natGateways)[i].PublicIps != nil {
				publicIpsList, serializeDiags = mappingNatGatewayPublicIps(ctx, natGateways, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			itemsValue[i], serializeDiags = datasource_nat_gateway.NewItemsValue(datasource_nat_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"id":         types.StringValue(utils.ConvertStringPtrToString((*natGateways)[i].Id)),
				"public_ips": publicIpsList,
				"state":      types.StringValue(utils.ConvertStringPtrToString((*natGateways)[i].State)),
				"subnet_id":  types.StringValue(utils.ConvertStringPtrToString((*natGateways)[i].SubnetId)),
				"tags":       tagsList,
				"vpc_id":     types.StringValue(utils.ConvertStringPtrToString((*natGateways)[i].VpcId)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		natGatewaysList, serializeDiags = types.ListValueFrom(ctx, new(datasource_nat_gateway.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		natGatewaysList = types.ListNull(new(datasource_nat_gateway.ItemsValue).Type(ctx))
	}

	return datasource_nat_gateway.NatGatewayModel{
		Items: natGatewaysList,
	}
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

func mappingNatGatewayTags(ctx context.Context, natGateways *[]api.NatGateway, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*natGateways)[i].Tags)
	elementValue := make([]datasource_nat_gateway.TagsValue, lt)
	for y, tag := range *(*natGateways)[i].Tags {
		elementValue[y], *diags = datasource_nat_gateway.NewTagsValue(datasource_nat_gateway.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_nat_gateway.TagsValue).Type(ctx), elementValue)
}

func mappingNatGatewayPublicIps(ctx context.Context, natGateways *[]api.NatGateway, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*natGateways)[i].PublicIps)
	elementValue := make([]datasource_nat_gateway.PublicIpsValue, lt)
	for y, publicIp := range *(*natGateways)[i].PublicIps {
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
