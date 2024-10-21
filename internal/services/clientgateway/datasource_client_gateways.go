package clientgateway

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type ClientGatewaysDataSourceModel struct {
	Items           []ClientGatewayModel `tfsdk:"items"`
	BgpAsns         types.List           `tfsdk:"bgp_asns"`
	ConnectionTypes types.List           `tfsdk:"connection_types"`
	IDs             types.List           `tfsdk:"ids"`
	PublicIps       types.List           `tfsdk:"public_ips"`
	States          types.List           `tfsdk:"states"`
	TagKeys         types.List           `tfsdk:"tag_keys"`
	TagValues       types.List           `tfsdk:"tag_values"`
	Tags            types.List           `tfsdk:"tags"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &clientGatewaysDataSource{}
)

func (d *clientGatewaysDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewClientGatewaysDataSource() datasource.DataSource {
	return &clientGatewaysDataSource{}
}

type clientGatewaysDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *clientGatewaysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_gateways"
}

// Schema defines the schema for the data source.
func (d *clientGatewaysDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = ClientGatewayDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *clientGatewaysDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan ClientGatewaysDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := ClientGatewaysFromTfToAPIReadParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}

	res := utils.ExecuteRequest(func() (*numspot.ReadClientGatewaysResponse, error) {
		return numspotClient.ReadClientGatewaysWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Client Gateways list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, clientGatewaysFromHttpToTfDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func ClientGatewaysFromTfToAPIReadParams(ctx context.Context, tf ClientGatewaysDataSourceModel, diags *diag.Diagnostics) numspot.ReadClientGatewaysParams {
	return numspot.ReadClientGatewaysParams{
		States:          utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		TagKeys:         utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:       utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:            utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:             utils.TfStringListToStringPtrList(ctx, tf.IDs, diags),
		ConnectionTypes: utils.TfStringListToStringPtrList(ctx, tf.ConnectionTypes, diags),
		BgpAsns:         utils.TFInt64ListToIntListPointer(ctx, tf.BgpAsns, diags),
		PublicIps:       utils.TfStringListToStringPtrList(ctx, tf.PublicIps, diags),
	}
}

func clientGatewaysFromHttpToTfDatasource(ctx context.Context, http *numspot.ClientGateway, diags *diag.Diagnostics) *ClientGatewayModel {
	var (
		tagsList types.List
		bgpAsnTf types.Int64
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	if http.BgpAsn != nil {
		bgpAsn := int64(*http.BgpAsn)
		bgpAsnTf = types.Int64PointerValue(&bgpAsn)
	}

	return &ClientGatewayModel{
		Id:             types.StringPointerValue(http.Id),
		State:          types.StringPointerValue(http.State),
		ConnectionType: types.StringPointerValue(http.ConnectionType),
		Tags:           tagsList,
		BgpAsn:         bgpAsnTf,
		PublicIp:       types.StringPointerValue(http.PublicIp),
	}
}
