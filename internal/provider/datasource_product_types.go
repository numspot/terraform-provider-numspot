package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_product_type"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type ProductTypesDataSourceModel struct {
	ProductTypes []datasource_product_type.ProductTypeModel `tfsdk:"product_types"`
	IDs          types.List                                 `tfsdk:"ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &productTypesDataSource{}
)

func (d *productTypesDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewProductTypesDataSource() datasource.DataSource {
	return &productTypesDataSource{}
}

type productTypesDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *productTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_product_types"
}

// Schema defines the schema for the data source.
func (d *productTypesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_product_type.ProductTypeDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *productTypesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan ProductTypesDataSourceModel
	request.Config.Get(ctx, &plan)

	params := ProductTypesFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadProductTypesResponse, error) {
		return d.provider.ApiClient.ReadProductTypesWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty ProductTypes list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := ProductTypesFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting ProductType HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.ProductTypes = append(state.ProductTypes, *tf)
	}
	state.IDs = plan.IDs

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
