package keypair

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/numspot/terraform-provider-numspot/internal/core"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/services"
	"github.com/numspot/terraform-provider-numspot/internal/services/keypair/datasource_keypair"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &keypairsDataSource{}

// keypairsDataSource is the Terraform data source implementation for numspot_keypairs.
type keypairsDataSource struct {
	services.BaseDataSource
}

// NewKeypairsDataSource returns a new keypairsDataSource instance.
func NewKeypairsDataSource() datasource.DataSource {
	return &keypairsDataSource{}
}

// Metadata sets the data source type name.
func (d *keypairsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_keypairs"
}

// Schema defines the schema for the data source.
func (d *keypairsDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_keypair.KeypairDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *keypairsDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_keypair.KeypairModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadKeypairs(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadKeypairs(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read keypairs", err.Error())
		return
	}

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *result, mappingItemsValue)
	if serializeDiags.HasError() {
		response.Diagnostics.Append(serializeDiags...)
		return
	}

	listValueItems := utils.CreateListValueItems(ctx, objectItems, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = listValueItems

	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func mappingItemsValue(ctx context.Context, item *api.Keypair) (datasource_keypair.ItemsValue, diag.Diagnostics) {
	value, valueDiags := datasource_keypair.NewItemsValue(datasource_keypair.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"fingerprint": types.StringValue(utils.ConvertStringPtrToString(item.Fingerprint)),
		"name":        types.StringValue(utils.ConvertStringPtrToString(item.Name)),
		"type":        types.StringValue(utils.ConvertStringPtrToString(item.Type)),
	})
	return value, valueDiags
}

func deserializeReadKeypairs(ctx context.Context, tf *datasource_keypair.KeypairModel, diags *diag.Diagnostics) *api.ReadKeypairsParams {
	return &api.ReadKeypairsParams{
		KeypairFingerprints: utils.ConvertTfListToArrayOfString(ctx, tf.KeypairFingerprints, diags),
		KeypairNames:        utils.ConvertTfListToArrayOfString(ctx, tf.KeypairNames, diags),
		KeypairTypes:        utils.ConvertTfListToArrayOfString(ctx, tf.KeypairTypes, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
