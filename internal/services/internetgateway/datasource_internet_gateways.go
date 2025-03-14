package internetgateway

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &internetGatewaysDataSource{}
)

func (d *internetGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewInternetGatewaysDataSource() datasource.DataSource {
	return &internetGatewaysDataSource{}
}

type internetGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *internetGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_internet_gateways"
}

// Schema defines the schema for the data source.
func (d *internetGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = InternetGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *internetGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan InternetGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	internetGatewayParams := deserializeReadInternetGateway(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	internetGateway, err := core.ReadInternetGatewaysWithParams(ctx, d.provider, internetGatewayParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read internet gateway", err.Error())
		return
	}

	internetGatewayItems := serializeInternetGateways(ctx, internetGateway, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = internetGatewayItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func serializeInternetGateways(ctx context.Context, internetGateways *[]numspot.InternetGateway, diags *diag.Diagnostics) []InternetGatewayModel {
	return utils.FromHttpGenericListToTfList(ctx, internetGateways, func(ctx context.Context, internetGateway *numspot.InternetGateway, diags *diag.Diagnostics) *InternetGatewayModel {
		var tagsList types.List

		if internetGateway.Tags != nil {
			tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *internetGateway.Tags, diags)
		}

		return &InternetGatewayModel{
			Id:    types.StringPointerValue(internetGateway.Id),
			State: types.StringPointerValue(internetGateway.State),
			VpcId: types.StringPointerValue(internetGateway.VpcId),
			Tags:  tagsList,
		}
	}, diags)
}

func deserializeReadInternetGateway(ctx context.Context, tf InternetGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadInternetGatewaysParams {
	return numspot.ReadInternetGatewaysParams{
		TagKeys:    utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:  utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:       utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:        utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		LinkStates: utils.TfStringListToStringPtrList(ctx, tf.LinkStates, diags),
		LinkVpcIds: utils.TfStringListToStringPtrList(ctx, tf.LinkVpcIds, diags),
	}
}
