package internetgateway

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
	"terraform-provider-numspot/internal/services/internetgateway/datasource_internet_gateway"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &internetGatewaysDataSource{}
)

func (d *internetGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewInternetGatewaysDataSource() datasource.DataSource {
	return &internetGatewaysDataSource{}
}

type internetGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *internetGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateways"
}

// Schema defines the schema for the data source.
func (d *internetGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_internet_gateway.InternetGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *internetGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetGatewayParams := deserializeReadInternetGateway(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	internetGateway, err := core.ReadInternetGatewaysWithParams(ctx, d.provider, internetGatewayParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read internet gateway", err.Error())
		return
	}

	internetGatewayItems := serializeInternetGateways(ctx, internetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = internetGatewayItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func serializeInternetGateways(ctx context.Context, internetGateways *[]api.InternetGateway, diags *diag.Diagnostics) datasource_internet_gateway.InternetGatewayModel {
	var internetGatewaysList types.List
	var serializeDiags diag.Diagnostics
	tagsList := types.List{}

	if len(*internetGateways) != 0 {
		ll := len(*internetGateways)
		itemsValue := make([]datasource_internet_gateway.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			if (*internetGateways)[i].Tags != nil {

				tagsList, serializeDiags = mappingInternetGatewayTags(ctx, internetGateways, diags, i)
				if serializeDiags.HasError() {
					diags.Append(serializeDiags...)
				}
			}

			itemsValue[i], serializeDiags = datasource_internet_gateway.NewItemsValue(datasource_internet_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"id":     types.StringValue(utils.ConvertStringPtrToString((*internetGateways)[i].Id)),
				"state":  types.StringValue(utils.ConvertStringPtrToString((*internetGateways)[i].State)),
				"tags":   tagsList,
				"vpc_id": types.StringValue(utils.ConvertStringPtrToString((*internetGateways)[i].VpcId)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		internetGatewaysList, serializeDiags = types.ListValueFrom(ctx, new(datasource_internet_gateway.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		internetGatewaysList = types.ListNull(new(datasource_internet_gateway.ItemsValue).Type(ctx))
	}

	return datasource_internet_gateway.InternetGatewayModel{
		Items: internetGatewaysList,
	}
}

func deserializeReadInternetGateway(ctx context.Context, tf datasource_internet_gateway.InternetGatewayModel, diags *diag.Diagnostics) api.ReadInternetGatewaysParams {
	return api.ReadInternetGatewaysParams{
		TagKeys:    utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:  utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:       utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:        utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		LinkStates: utils.ConvertTfListToArrayOfString(ctx, tf.LinkStates, diags),
		LinkVpcIds: utils.ConvertTfListToArrayOfString(ctx, tf.LinkVpcIds, diags),
	}
}

func mappingInternetGatewayTags(ctx context.Context, internetGateways *[]api.InternetGateway, diags *diag.Diagnostics, i int) (types.List, diag.Diagnostics) {
	lt := len(*(*internetGateways)[i].Tags)
	elementValue := make([]datasource_internet_gateway.TagsValue, lt)
	for y, tag := range *(*internetGateways)[i].Tags {
		elementValue[y], *diags = datasource_internet_gateway.NewTagsValue(datasource_internet_gateway.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
			"key":   types.StringValue(tag.Key),
			"value": types.StringValue(tag.Value),
		})
		if diags.HasError() {
			diags.Append(*diags...)
			continue
		}
	}

	return types.ListValueFrom(ctx, new(datasource_internet_gateway.TagsValue).Type(ctx), elementValue)
}
