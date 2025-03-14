package publicip

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

type PublicIpsDataSourceModel struct {
	Items           []PublicIpModelDatasource `tfsdk:"items"`
	LinkPublicIpIds types.List                `tfsdk:"link_public_ip_ids"`
	NicIds          types.List                `tfsdk:"nic_ids"`
	TagKeys         types.List                `tfsdk:"tag_keys"`
	TagValues       types.List                `tfsdk:"tag_values"`
	Tags            types.List                `tfsdk:"tags"`
	PrivateIps      types.List                `tfsdk:"private_ips"`
	VmIds           types.List                `tfsdk:"vm_ids"`
	IDs             types.List                `tfsdk:"ids"`
}

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
			"Unexpected Resource Configure Type",
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
	resp.Schema = PublicIpDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *publicIpsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan PublicIpsDataSourceModel
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
		response.Diagnostics.AddError("unable to read internet gateway", err.Error())
		return
	}

	objectItems := serializePublicIps(ctx, numSpotPublicIp, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializePublicIpParams(ctx context.Context, tf PublicIpsDataSourceModel, diags *diag.Diagnostics) numspot.ReadPublicIpsParams {
	return numspot.ReadPublicIpsParams{
		LinkPublicIpIds: utils.TfStringListToStringPtrList(ctx, tf.LinkPublicIpIds, diags),
		NicIds:          utils.TfStringListToStringPtrList(ctx, tf.NicIds, diags),
		PrivateIps:      utils.TfStringListToStringPtrList(ctx, tf.PrivateIps, diags),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		VmIds:           utils.TfStringListToStringPtrList(ctx, tf.VmIds, diags),
	}
}

func serializePublicIps(ctx context.Context, publicIp *[]numspot.PublicIp, diags *diag.Diagnostics) []PublicIpModelDatasource {
	return utils.FromHttpGenericListToTfList(ctx, publicIp, func(ctx context.Context, publicIP *numspot.PublicIp, diags *diag.Diagnostics) *PublicIpModelDatasource {
		var tagsList types.List

		if publicIP.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *publicIP.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		return &PublicIpModelDatasource{
			Id:             types.StringPointerValue(publicIP.Id),
			NicId:          types.StringPointerValue(publicIP.NicId),
			PrivateIp:      types.StringPointerValue(publicIP.PrivateIp),
			PublicIp:       types.StringPointerValue(publicIP.PublicIp),
			VmId:           types.StringPointerValue(publicIP.VmId),
			LinkPublicIpId: types.StringPointerValue(publicIP.LinkPublicIpId),
			Tags:           tagsList,
		}
	}, diags)
}
