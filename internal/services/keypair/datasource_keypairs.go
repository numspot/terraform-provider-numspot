package keypair

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/core"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

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

	keypairParams := deserializeReadKeypairs(ctx, plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	keypair, err := core.ReadKeypairs(ctx, d.provider, keypairParams)
	if err != nil {
		response.Diagnostics.AddError("unable to read keypairs", err.Error())
		return
	}

	keypairItems := serializeKeypairs(ctx, keypair, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = keypairItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func serializeKeypairs(ctx context.Context, keypairs *[]numspot.Keypair, diags *diag.Diagnostics) []KeyPairDatasourceItemModel {
	return utils.FromHttpGenericListToTfList(ctx, keypairs, func(ctx context.Context, http *numspot.Keypair, diags *diag.Diagnostics) *KeyPairDatasourceItemModel {
		return &KeyPairDatasourceItemModel{
			Fingerprint: types.StringPointerValue(http.Fingerprint),
			Name:        types.StringPointerValue(http.Name),
			Type:        types.StringPointerValue(http.Type),
		}
	}, diags)
}

func deserializeReadKeypairs(ctx context.Context, tf KeypairsDataSourceModel, diags *diag.Diagnostics) numspot.ReadKeypairsParams {
	return numspot.ReadKeypairsParams{
		KeypairFingerprints: utils.TfStringListToStringPtrList(ctx, tf.KeypairFingerprints, diags),
		KeypairNames:        utils.TfStringListToStringPtrList(ctx, tf.KeypairNames, diags),
		KeypairTypes:        utils.TfStringListToStringPtrList(ctx, tf.KeypairTypes, diags),
	}
}
