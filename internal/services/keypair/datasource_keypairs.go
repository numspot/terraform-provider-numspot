package keypair

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
	"terraform-provider-numspot/internal/services/keypair/datasource_keypair"
	"terraform-provider-numspot/internal/utils"
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
			"Unexpected Datasource Configure Type",
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
	resp.Schema = datasource_keypair.KeypairDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *keypairsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_keypair.KeypairModel

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

	keypairItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *keypair, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, keypairItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, keypair api.Keypair) (datasource_keypair.ItemsValue, diag.Diagnostics) {
	return datasource_keypair.NewItemsValue(datasource_keypair.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"name":        types.StringValue(utils.ConvertStringPtrToString(keypair.Name)),
		"fingerprint": types.StringValue(utils.ConvertStringPtrToString(keypair.Fingerprint)),
		"type":        types.StringValue(utils.ConvertStringPtrToString(keypair.Type)),
	})
}

func deserializeReadKeypairs(ctx context.Context, tf datasource_keypair.KeypairModel, diags *diag.Diagnostics) api.ReadKeypairsParams {
	return api.ReadKeypairsParams{
		KeypairFingerprints: utils.ConvertTfListToArrayOfString(ctx, tf.KeypairFingerprints, diags),
		KeypairNames:        utils.ConvertTfListToArrayOfString(ctx, tf.KeypairNames, diags),
		KeypairTypes:        utils.ConvertTfListToArrayOfString(ctx, tf.KeypairTypes, diags),
	}
}
