package vpc

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

type VPCsDataSourceModel struct {
	Items             []VpcModel `tfsdk:"items"`
	IDs               types.List `tfsdk:"ids"`
	DHCPOptionsSetIds types.List `tfsdk:"dhcp_options_set_ids"`
	IPRanges          types.List `tfsdk:"ip_ranges"`
	IsDefault         types.Bool `tfsdk:"is_default"`
	States            types.List `tfsdk:"states"`
	TagKeys           types.List `tfsdk:"tag_keys"`
	TagValues         types.List `tfsdk:"tag_values"`
	Tags              types.List `tfsdk:"tags"`
}

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
			"Unexpected Resource Configure Type",
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
	resp.Schema = VpcDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *vpcsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan VPCsDataSourceModel
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

	objectItems := serializeVPCs(ctx, numSpotVpc, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeVPCParams(ctx context.Context, tf VPCsDataSourceModel, diags *diag.Diagnostics) numspot.ReadVpcsParams {
	return numspot.ReadVpcsParams{
		DhcpOptionsSetIds: utils.TfStringListToStringPtrList(ctx, tf.DHCPOptionsSetIds, diags),
		IpRanges:          utils.TfStringListToStringPtrList(ctx, tf.IPRanges, diags),
		IsDefault:         tf.IsDefault.ValueBoolPointer(),
		States:            utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:           utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:         utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:              utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:               utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
	}
}

func serializeVPCs(ctx context.Context, vpcs *[]numspot.Vpc, diags *diag.Diagnostics) []VpcModel {
	return utils.FromHttpGenericListToTfList(ctx, vpcs, func(ctx context.Context, http *numspot.Vpc, diags *diag.Diagnostics) *VpcModel {
		var tagsList types.List

		if http.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		}

		return &VpcModel{
			DhcpOptionsSetId: types.StringPointerValue(http.DhcpOptionsSetId),
			Id:               types.StringPointerValue(http.Id),
			IpRange:          types.StringPointerValue(http.IpRange),
			State:            types.StringPointerValue(http.State),
			Tenancy:          types.StringPointerValue(http.Tenancy),
			Tags:             tagsList,
		}
	}, diags)
}
