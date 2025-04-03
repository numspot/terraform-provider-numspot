package subnet

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
	"terraform-provider-numspot/internal/services/subnet/datasource_subnet"
	"terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &subnetsDataSource{}
)

func (d *subnetsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewSubnetsDataSource() datasource.DataSource {
	return &subnetsDataSource{}
}

type subnetsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *subnetsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subnets"
}

// Schema defines the schema for the data source.
func (d *subnetsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_subnet.SubnetDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *subnetsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_subnet.SubnetModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numspotSubnet, err := core.ReadSubnetsWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read subnets", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *numspotSubnet, mappingItemsValue)
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

func deserializeParams(ctx context.Context, tf datasource_subnet.SubnetModel, diags *diag.Diagnostics) api.ReadSubnetsParams {
	return api.ReadSubnetsParams{
		AvailableIpsCounts:    utils.ConvertTfListToArrayOfInt(ctx, tf.AvailableIpsCounts, diags),
		IpRanges:              utils.ConvertTfListToArrayOfString(ctx, tf.IpRanges, diags),
		States:                utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		VpcIds:                utils.ConvertTfListToArrayOfString(ctx, tf.VpcIds, diags),
		Ids:                   utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
	}
}

func mappingItemsValue(ctx context.Context, subnet api.Subnet) (datasource_subnet.ItemsValue, diag.Diagnostics) {
	tagsList := types.ListNull(datasource_subnet.ItemsValue{}.Type(ctx))

	if subnet.Tags != nil {
		tagItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *subnet.Tags, mappingTags)
		if serializeDiags.HasError() {
			return datasource_subnet.ItemsValue{}, serializeDiags
		}
		tagsList = utils.CreateListValueItems(ctx, tagItems, &serializeDiags)
		if serializeDiags.HasError() {
			return datasource_subnet.ItemsValue{}, serializeDiags
		}
	}

	return datasource_subnet.NewItemsValue(datasource_subnet.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name":  types.StringValue(utils.ConvertAzNamePtrToString(subnet.AvailabilityZoneName)),
		"available_ips_count":     types.Int64Value(utils.ConvertIntPtrToInt64(subnet.AvailableIpsCount)),
		"id":                      types.StringValue(utils.ConvertStringPtrToString(subnet.Id)),
		"ip_range":                types.StringValue(utils.ConvertStringPtrToString(subnet.IpRange)),
		"map_public_ip_on_launch": types.BoolPointerValue(subnet.MapPublicIpOnLaunch),
		"state":                   types.StringValue(utils.ConvertStringPtrToString(subnet.State)),
		"tags":                    tagsList,
		"vpc_id":                  types.StringValue(utils.ConvertStringPtrToString(subnet.VpcId)),
	})
}

func mappingTags(ctx context.Context, tag api.ResourceTag) (datasource_subnet.TagsValue, diag.Diagnostics) {
	return datasource_subnet.NewTagsValue(datasource_subnet.TagsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"key":   types.StringValue(tag.Key),
		"value": types.StringValue(tag.Value),
	})
}
