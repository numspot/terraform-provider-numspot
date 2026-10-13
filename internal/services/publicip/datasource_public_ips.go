package publicip

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/publicip/datasource_public_ip"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &publicipsDataSource{}

// publicipsDataSource is the Terraform data source implementation for numspot_public_ips.
type publicipsDataSource struct {
	services.BaseDataSource
}

// NewPublicIpsDataSource returns a new publicipsDataSource instance.
func NewPublicIpsDataSource() datasource.DataSource {
	return &publicipsDataSource{}
}

// Metadata sets the data source type name.
func (d *publicipsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ips"
}

// Schema defines the schema for the data source.
func (d *publicipsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_public_ip.PublicIpDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *publicipsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_public_ip.PublicIpModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadPublicIps(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadPublicIpsWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read public_ips", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.PublicIp) (datasource_public_ip.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_public_ip.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_public_ip.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_public_ip.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_public_ip.NewItemsValue(datasource_public_ip.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"id":                types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"link_public_ip_id": types.StringValue(utils.ConvertStringPtrToString(item.LinkPublicIpId)),
		"nic_id":            types.StringValue(utils.ConvertStringPtrToString(item.NicId)),
		"private_ip":        types.StringValue(utils.ConvertStringPtrToString(item.PrivateIp)),
		"public_ip":         types.StringValue(utils.ConvertStringPtrToString(item.PublicIp)),
		"vm_id":             types.StringValue(utils.ConvertStringPtrToString(item.VmId)),
		"tags":              tagsList,
	})
	return value, valueDiags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_public_ip.TagsValue, diag.Diagnostics) {
	return datasource_public_ip.NewTagsValue(datasource_public_ip.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func deserializeReadPublicIps(ctx context.Context, tf *datasource_public_ip.PublicIpModel, diags *diag.Diagnostics) *api.ReadPublicIpsParams {
	return &api.ReadPublicIpsParams{
		LinkPublicIpIds: utils.ConvertTfListToArrayOfString(ctx, tf.LinkPublicIpIds, diags),
		NicIds:          utils.ConvertTfListToArrayOfString(ctx, tf.NicIds, diags),
		PrivateIps:      utils.ConvertTfListToArrayOfString(ctx, tf.PrivateIps, diags),
		TagKeys:         utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:       utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:            utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		VmIds:           utils.ConvertTfListToArrayOfString(ctx, tf.VmIds, diags),
		Ids:             utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
