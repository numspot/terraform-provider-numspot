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

	keypairItems := serializeKeypairs(ctx, keypair, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = keypairItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func serializeKeypairs(ctx context.Context, keypairs *[]api.Keypair, diags *diag.Diagnostics) datasource_keypair.KeypairModel {
	var keypairsList types.List
	var serializeDiags diag.Diagnostics

	if len(*keypairs) != 0 {
		ll := len(*keypairs)
		itemsValue := make([]datasource_keypair.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = datasource_keypair.NewItemsValue(datasource_keypair.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"name":        types.StringValue(utils.ConvertStringPtrToString((*keypairs)[i].Name)),
				"fingerprint": types.StringValue(utils.ConvertStringPtrToString((*keypairs)[i].Fingerprint)),
				"type":        types.StringValue(utils.ConvertStringPtrToString((*keypairs)[i].Type)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		keypairsList, serializeDiags = types.ListValueFrom(ctx, new(datasource_keypair.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		keypairsList = types.ListNull(new(datasource_keypair.ItemsValue).Type(ctx))
	}

	return datasource_keypair.KeypairModel{
		Items: keypairsList,
	}
}

func deserializeReadKeypairs(ctx context.Context, tf datasource_keypair.KeypairModel, diags *diag.Diagnostics) api.ReadKeypairsParams {
	return api.ReadKeypairsParams{
		KeypairFingerprints: utils.ConvertTfListToArrayOfString(ctx, tf.KeypairFingerprints, diags),
		KeypairNames:        utils.ConvertTfListToArrayOfString(ctx, tf.KeypairNames, diags),
		KeypairTypes:        utils.ConvertTfListToArrayOfString(ctx, tf.KeypairTypes, diags),
	}
}
