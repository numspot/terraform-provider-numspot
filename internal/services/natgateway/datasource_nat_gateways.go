package natgateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type NatGatewaysDataSourceModel struct {
	Items     []NatGatewayModelDatasource `tfsdk:"items"`
	Ids       types.List                  `tfsdk:"ids"`
	States    types.List                  `tfsdk:"states"`
	SubnetIds types.List                  `tfsdk:"subnet_ids"`
	TagKeys   types.List                  `tfsdk:"tag_keys"`
	TagValues types.List                  `tfsdk:"tag_values"`
	Tags      types.List                  `tfsdk:"tags"`
	VpcIds    types.List                  `tfsdk:"vpc_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &natGatewaysDataSource{}
)

func (d *natGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewNatGatewaysDataSource() datasource.DataSource {
	return &natGatewaysDataSource{}
}

type natGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *natGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_nat_gateways"
}

// Schema defines the schema for the data source.
func (d *natGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = NatGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *natGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan NatGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := deserializeNatGatewaysParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numSpotNATGateway, err := core.ReadNATGatewaysWithParams(ctx, d.provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read nat gateway", err.Error())
		return
	}

	objectItems := serializeNATGateways(ctx, numSpotNATGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func serializeNATGateways(ctx context.Context, natGateways *[]numspot.NatGateway, diags *diag.Diagnostics) []NatGatewayModelDatasource {
	return utils.FromHttpGenericListToTfList(ctx, natGateways, func(ctx context.Context, http *numspot.NatGateway, diags *diag.Diagnostics) *NatGatewayModelDatasource {
		var tagsTf types.List

		var publicIp []numspot.PublicIpLight
		if http.PublicIps != nil {
			publicIp = *http.PublicIps
		}
		// Public Ips
		publicIpsTf := utils.GenericListToTfListValue(
			ctx,
			serializePublicIp,
			publicIp,
			diags,
		)
		if diags.HasError() {
			return nil
		}

		if http.Tags != nil {
			tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
			if diags.HasError() {
				return nil
			}
		}

		return &NatGatewayModelDatasource{
			Id:        types.StringPointerValue(http.Id),
			PublicIps: publicIpsTf,
			State:     types.StringPointerValue(http.State),
			SubnetId:  types.StringPointerValue(http.SubnetId),
			VpcId:     types.StringPointerValue(http.VpcId),
			Tags:      tagsTf,
		}
	}, diags)
}

func deserializeNatGatewaysParams(ctx context.Context, tf NatGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadNatGatewayParams {
	return numspot.ReadNatGatewayParams{
		SubnetIds: utils.TfStringListToStringPtrList(ctx, tf.SubnetIds, diags),
		VpcIds:    utils.TfStringListToStringPtrList(ctx, tf.VpcIds, diags),
		States:    utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:   utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues: utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:      utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:       utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
	}
}
