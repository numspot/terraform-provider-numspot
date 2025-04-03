package publicip

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
	"terraform-provider-numspot/internal/services/publicip/datasource_public_ip"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &publicIpsDataSource{}
)

func (d *publicIpsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewPublicIpsDataSource() datasource.DataSource {
	return &publicIpsDataSource{}
}

type publicIpsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *publicIpsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ips"
}

// Schema defines the schema for the data source.
func (d *publicIpsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_public_ip.PublicIpDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *publicIpsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializePublicIpParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotPublicIp, err := core.ReadPublicIpsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read public ip", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *numSpotPublicIp, mappingItemsValue)
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

func deserializePublicIpParams(ctx context.Context, tf datasource_public_ip.PublicIpModel, diags *diag.Diagnostics) api.ReadPublicIpsParams {
	return api.ReadPublicIpsParams{
		LinkPublicIpIds: utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpIds, diags),
		NicIds:          utils.ConvertTfListToArrayOfString(ctx, tf.NicIds, diags),
		PrivateIps:      utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIps, diags),
		TagKeys:         utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:       utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:            utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:             utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		VmIds:           utils.ConvertTfListToArrayOfString(ctx, tf.VmIds, diags),
	}
}

func mappingItemsValue(ctx context.Context, publicIp api.PublicIp) (datasource_public_ip.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_public_ip.ItemsValue{}.Type(ctx))

	if publicIp.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *publicIp.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_public_ip.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_public_ip.ItemsValue{}, serializeDiags
		}
	}

	return datasource_public_ip.NewItemsValue(datasource_public_ip.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":                types.StringValue(utils.ConvertStringPtrToString(publicIp.Id)),
		"link_public_ip_id": types.StringValue(utils.ConvertStringPtrToString(publicIp.LinkPublicIpId)),
		"nic_id":            types.StringValue(utils.ConvertStringPtrToString(publicIp.NicId)),
		"private_ip":        types.StringValue(utils.ConvertStringPtrToString(publicIp.PrivateIp)),
		"public_ip":         types.StringValue(utils.ConvertStringPtrToString(publicIp.PublicIp)),
		"tags":              tagsList,
		"vm_id":             types.StringValue(utils.ConvertStringPtrToString(publicIp.VmId)),
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_public_ip.TagsValue, diag.Diagnostics) {
	return datasource_public_ip.NewTagsValue(datasource_public_ip.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
