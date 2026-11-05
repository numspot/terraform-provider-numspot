package vpc

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/vpc/datasource_vpc"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &vpcsDataSource{}

// vpcsDataSource is the Terraform data source implementation for numspot_vpcs.
type vpcsDataSource struct {
	services.BaseDataSource
}

// NewVPCsDataSource returns a new vpcsDataSource instance.
func NewVPCsDataSource() datasource.DataSource {
	return &vpcsDataSource{}
}

// Metadata sets the data source type name.
func (d *vpcsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpcs"
}

// Schema defines the schema for the data source.
func (d *vpcsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_vpc.VpcDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *vpcsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_vpc.VpcModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadVPCs(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadVPCsWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read vpcs", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.Vpc) (datasource_vpc.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_vpc.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_vpc.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_vpc.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_vpc.NewItemsValue(datasource_vpc.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"dhcp_options_set_id": types.StringValue(utils.ConvertStringPtrToString(item.DhcpOptionsSetId)),
		"id":                  types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"ip_range":            types.StringValue(utils.ConvertStringPtrToString(item.IpRange)),
		"state":               types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"tenancy":             types.StringValue(utils.ConvertStringPtrToString(item.Tenancy)),
		"tags":                tagsList,
	})
	return value, valueDiags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_vpc.TagsValue, diag.Diagnostics) {
	return datasource_vpc.NewTagsValue(datasource_vpc.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func deserializeReadVPCs(ctx context.Context, tf *datasource_vpc.VpcModel, diags *diag.Diagnostics) *api.ReadVpcsParams {
	return &api.ReadVpcsParams{
		DhcpOptionsSetIds: utils.ConvertTfListToArrayOfString(ctx, tf.DhcpOptionsSetIds, diags),
		IpRanges:          utils.ConvertTfListToArrayOfString(ctx, tf.IpRanges, diags),
		IsDefault:         tf.IsDefault.ValueBoolPointer(),
		States:            utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		TagKeys:           utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:         utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:              utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		Ids:               utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
