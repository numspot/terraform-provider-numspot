package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_key_pair"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type KeypairDataSourceModel struct {
	Keypair      []datasource_key_pair.KeyPairModel `tfsdk:"keypairs"`
	Fingerprints types.List                         `tfsdk:"fingerprints"`
	Names        types.List                         `tfsdk:"names"`
	Types        types.List                         `tfsdk:"types"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &KeypairDataSource{}
)

func (d *KeypairDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
	if request.ProviderData == nil {
		return
	}

	provider, ok := request.ProviderData.(Provider)
	if !ok {
		response.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
		)

		return
	}

	d.provider = provider
}

func NewKeypairDataSource() datasource.DataSource {
	return &KeypairDataSource{}
}

type KeypairDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *KeypairDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypair"
}

// Schema defines the schema for the data source.
func (d *KeypairDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_key_pair.KeyPairDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *KeypairDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan KeypairDataSourceModel
	request.Config.Get(ctx, &plan)

	params := KeypairFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadKeypairsResponse, error) {
		return d.provider.ApiClient.ReadKeypairsWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty Keypair list")
	}

	for _, item := range *res.JSON200.Items {
		tf := KeypairFromHttpToTfDatasource(ctx, &item)
		state.Keypair = append(state.Keypair, *tf)
	}

	state.Fingerprints = plan.Fingerprints
	state.Names = plan.Names
	state.Types = plan.Types

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
