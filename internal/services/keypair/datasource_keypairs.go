package keypair

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type KeypairsDataSourceModel struct {
	Items               []KeyPairDatasourceItemModel `tfsdk:"items"`
	KeypairFingerprints types.List                   `tfsdk:"keypair_fingerprints"`
	KeypairNames        types.List                   `tfsdk:"keypair_names"`
	KeypairTypes        types.List                   `tfsdk:"keypair_types"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &keypairsDataSource{}
)

func (d *keypairsDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewKeypairsDataSource() datasource.DataSource {
	return &keypairsDataSource{}
}

type keypairsDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *keypairsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypairs"
}

// Schema defines the schema for the data source.
func (d *keypairsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = KeyPairDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *keypairsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan KeypairsDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	params := KeypairsFromTfToAPIReadParams(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}
	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	res := utils.ExecuteRequest(func() (*numspot.ReadKeypairsResponse, error) {
		return numspotClient.ReadKeypairsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Keypair list")
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, KeypairsFromHttpToTfDatasource, &response.Diagnostics)

	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
