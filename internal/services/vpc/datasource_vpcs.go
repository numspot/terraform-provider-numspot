package vpc

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
	"terraform-provider-numspot/internal/services/vpc/datasource_vpc"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &vpcsDataSource{}
)

func (d *vpcsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewVPCsDataSource() datasource.DataSource {
	return &vpcsDataSource{}
}

type vpcsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *vpcsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

// Schema defines the schema for the data source.
func (d *vpcsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpc.VpcDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpcsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_vpc.VpcModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeVPCParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotVpc, err := core.ReadVPCsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read internet gateway", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *numSpotVpc, mappingItemsValue)
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

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeVPCParams(ctx context.Context, tf datasource_vpc.VpcModel, diags *diag.Diagnostics) api.ReadVpcsParams {
	return api.ReadVpcsParams{
		DhcpOptionsSetIds: utils.ConvertTfListToArrayOfString(ctx, tf.DhcpOptionsSetIds, diags),
		IpRanges:          utils.ConvertTfListToArrayOfString(ctx, tf.IpRanges, diags),
		IsDefault:         tf.IsDefault.ValueBoolPointer(),
		States:            utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		TagKeys:           utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:         utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:              utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:               utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
	}
}

func mappingItemsValue(ctx context.Context, vpc api.Vpc) (datasource_vpc.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_vpc.ItemsValue{}.Type(ctx))

	if vpc.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *vpc.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_vpc.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_vpc.ItemsValue{}, serializeDiags
		}
	}

	return datasource_vpc.NewItemsValue(datasource_vpc.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"dhcp_options_set_id": types.StringValue(utils.ConvertStringPtrToString(vpc.DhcpOptionsSetId)),
		"id":                  types.StringValue(utils.ConvertStringPtrToString(vpc.Id)),
		"ip_range":            types.StringValue(utils.ConvertStringPtrToString(vpc.IpRange)),
		"state":               types.StringValue(utils.ConvertStringPtrToString(vpc.State)),
		"tags":                tagsList,
		"tenancy":             types.StringValue(utils.ConvertStringPtrToString(vpc.Tenancy)),
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_vpc.TagsValue, diag.Diagnostics) {
	return datasource_vpc.NewTagsValue(datasource_vpc.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
