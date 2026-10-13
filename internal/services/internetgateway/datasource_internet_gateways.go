package internetgateway

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/internetgateway/datasource_internet_gateway"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &internetgatewaysDataSource{}

// internetgatewaysDataSource is the Terraform data source implementation for numspot_internet_gateways.
type internetgatewaysDataSource struct {
	services.BaseDataSource
}

// NewInternetGatewaysDataSource returns a new internetgatewaysDataSource instance.
func NewInternetGatewaysDataSource() datasource.DataSource {
	return &internetgatewaysDataSource{}
}

// Metadata sets the data source type name.
func (d *internetgatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateways"
}

// Schema defines the schema for the data source.
func (d *internetgatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_internet_gateway.InternetGatewayDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *internetgatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_internet_gateway.InternetGatewayModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadInternetGateways(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadInternetGatewaysWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read internet_gateways", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.InternetGateway) (datasource_internet_gateway.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_internet_gateway.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_internet_gateway.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_internet_gateway.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_internet_gateway.NewItemsValue(
		datasource_internet_gateway.ItemsValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"id":     types.StringValue(utils.ConvertStringPtrToString(item.Id)),
			"state":  types.StringValue(utils.ConvertStringPtrToString(item.State)),
			"vpc_id": types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
			"tags":   tagsList,
		},
	)
	return value, valueDiags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_internet_gateway.TagsValue, diag.Diagnostics) {
	return datasource_internet_gateway.NewTagsValue(datasource_internet_gateway.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func deserializeReadInternetGateways(
	ctx context.Context,
	tf *datasource_internet_gateway.InternetGatewayModel,
	diags *diag.Diagnostics,
) *api.ReadInternetGatewaysParams {
	return &api.ReadInternetGatewaysParams{
		LinkStates: utils.ConvertTfListToArrayOfString(ctx, tf.LinkStates, diags),
		TagKeys:    utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:  utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:       utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		LinkVpcIds: utils.ConvertTfListToArrayOfString(ctx, tf.LinkVpcIds, diags),
		Ids:        utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
