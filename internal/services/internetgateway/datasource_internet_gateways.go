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
	"terraform-provider-numspot/internal/services/vpc/datasource_vpc"
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
			"Unexpected Datasource Configure Type",
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

	internetGateways, err := core.ReadInternetGatewaysWithParams(ctx, d.provider, internetGatewayParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read internet gateway", err.Error())
		return
	}

	internetGatewayItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *internetGateways, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}
	listValueItems := utils.CreateListValueItems(ctx, internetGatewayItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
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

func mappingItemsValue(ctx context.Context, internetGateway api.InternetGateway) (datasource_internet_gateway.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_vpc.ItemsValue{}.Type(ctx))

	if internetGateway.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *internetGateway.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_internet_gateway.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_internet_gateway.ItemsValue{}, serializeDiags
		}
	}

	return datasource_internet_gateway.NewItemsValue(datasource_internet_gateway.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":     types.StringValue(utils.ConvertStringPtrToString(internetGateway.Id)),
		"state":  types.StringValue(utils.ConvertStringPtrToString(internetGateway.State)),
		"tags":   tagsList,
		"vpc_id": types.StringValue(utils.ConvertStringPtrToString(internetGateway.VpcId)),
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_internet_gateway.TagsValue, diag.Diagnostics) {
	return datasource_internet_gateway.NewTagsValue(datasource_internet_gateway.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
