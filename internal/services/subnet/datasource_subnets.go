package subnet

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/core"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

type SubnetsDataSourceModel struct {
	Items                 []SubnetModel `tfsdk:"items"`
	AvailabilityZoneNames types.List    `tfsdk:"availability_zone_names"`
	AvailableIpsCounts    types.List    `tfsdk:"available_ips_counts"`
	Ids                   types.List    `tfsdk:"ids"`
	IpRanges              types.List    `tfsdk:"ip_ranges"`
	States                types.List    `tfsdk:"states"`
	TagKeys               types.List    `tfsdk:"tag_keys"`
	TagValues             types.List    `tfsdk:"tag_values"`
	Tags                  types.List    `tfsdk:"tags"`
	VpcIds                types.List    `tfsdk:"vpc_ids"`
}

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
	resp.Schema = SubnetDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *subnetsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan SubnetsDataSourceModel
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

	objectItems := serializeSubnets(ctx, numspotSubnet, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeParams(ctx context.Context, tf SubnetsDataSourceModel, diags *diag.Diagnostics) numspot.ReadSubnetsParams {
	return numspot.ReadSubnetsParams{
		AvailableIpsCounts:    utils.TFInt64ListToIntListPointer(ctx, tf.AvailableIpsCounts, diags),
		IpRanges:              utils.TfStringListToStringPtrList(ctx, tf.IpRanges, diags),
		States:                utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		VpcIds:                utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
	}
}

func serializeSubnets(ctx context.Context, subnets *[]numspot.Subnet, diags *diag.Diagnostics) []SubnetModel {
	return utils.FromHttpGenericListToTfList(ctx, subnets, func(ctx context.Context, subnet *numspot.Subnet, diags *diag.Diagnostics) *SubnetModel {
		var tagsList types.List

		if subnet.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *subnet.Tags, diags)
		}

		return &SubnetModel{
			AvailabilityZoneName: types.StringPointerValue(subnet.AvailabilityZoneName),
			AvailableIpsCount:    utils.FromIntPtrToTfInt64(subnet.AvailableIpsCount),
			Id:                   types.StringPointerValue(subnet.Id),
			IpRange:              types.StringPointerValue(subnet.IpRange),
			MapPublicIpOnLaunch:  types.BoolPointerValue(subnet.MapPublicIpOnLaunch),
			State:                types.StringPointerValue(subnet.State),
			VpcId:                types.StringPointerValue(subnet.VpcId),
			Tags:                 tagsList,
		}
	}, diags)
}
