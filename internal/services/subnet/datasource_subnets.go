package subnet

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/subnet/datasource_subnet"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &subnetsDataSource{}

// subnetsDataSource is the Terraform data source implementation for numspot_subnets.
type subnetsDataSource struct {
	services.BaseDataSource
}

// NewSubnetsDataSource returns a new subnetsDataSource instance.
func NewSubnetsDataSource() datasource.DataSource {
	return &subnetsDataSource{}
}

// Metadata sets the data source type name.
func (d *subnetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

// Schema defines the schema for the data source.
func (d *subnetsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_subnet.SubnetDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *subnetsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_subnet.SubnetModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadSubnets(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadSubnetsWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read subnets", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.Subnet) (datasource_subnet.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_subnet.ItemsValue{}.Type(ctx))

	if item.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *item.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_subnet.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_subnet.ItemsValue{}, serializeDiags
		}
	}

	value, valueDiags := datasource_subnet.NewItemsValue(datasource_subnet.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name":  types.StringValue(utils.ConvertAzNamePtrToString(item.AvailabilityZoneName)),
		"available_ips_count":     utils.FromIntPtrToTfInt64(item.AvailableIpsCount),
		"id":                      types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"ip_range":                types.StringValue(utils.ConvertStringPtrToString(item.IpRange)),
		"map_public_ip_on_launch": types.BoolValue(*item.MapPublicIpOnLaunch),
		"state":                   types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"vpc_id":                  types.StringValue(utils.ConvertStringPtrToString(item.VpcId)),
		"tags":                    tagsList,
	})
	return value, valueDiags
}

func mappingTags(ctx context.Context, tag *api.ResourceTag) (datasource_subnet.TagsValue, diag.Diagnostics) {
	return datasource_subnet.NewTagsValue(datasource_subnet.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}

func deserializeReadSubnets(ctx context.Context, tf *datasource_subnet.SubnetModel, diags *diag.Diagnostics) *api.ReadSubnetsParams {
	return &api.ReadSubnetsParams{
		AvailableIpsCounts:    utils.ConvertTfListToArrayOfInt(ctx, tf.AvailableIpsCounts, diags),
		IpRanges:              utils.ConvertTfListToArrayOfString(ctx, tf.IpRanges, diags),
		States:                utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		TagKeys:               utils.ConvertTfListToArrayOfString(ctx, tf.TagKeys, diags),
		TagValues:             utils.ConvertTfListToArrayOfString(ctx, tf.TagValues, diags),
		Tags:                  utils.ConvertTfListToArrayOfString(ctx, tf.Tags, diags),
		VpcIds:                utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		Ids:                   utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
