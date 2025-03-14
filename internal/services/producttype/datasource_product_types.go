package producttype

//// Product Types are not handled for now
//
//import (
//	"context"
//	"fmt"
//	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/services"
//	"net/http"
//
//	"github.com/hashicorp/terraform-plugin-framework/datasource"
//	"github.com/hashicorp/terraform-plugin-framework/types"
//
//	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
//)
//
//type ProductTypesDataSourceModel struct {
//	Items []ProductTypeModel `tfsdk:"items"`
//	IDs   types.List         `tfsdk:"ids"`
//}
//
//// Ensure the implementation satisfies the expected interfaces.
//var (
//	_ datasource.DataSource = &productTypesDataSource{}
//)
//
//func (d *productTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
//	if req.ProviderData == nil {
//		return
//	}
//
//	provider, ok := request.ProviderData.(*client.NumSpotSDK)
//	if !ok {
//		response.Diagnostics.AddError(
//			"Unexpected Resource Configure Type",
//			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", request.ProviderData),
//		)
//
//		return
//	}
//
//	d.provider = provider
//}
//
//func NewProductTypesDataSource() datasource.DataSource {
//	return &productTypesDataSource{}
//}
//
//type productTypesDataSource struct {
//	provider *client.NumSpotSDK
//}
//
//// Metadata returns the data source type name.
//func (d *productTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
//	resp.TypeName = req.ProviderTypeName + "_product_types"
//}
//
//// Schema defines the schema for the data source.
//func (d *productTypesDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
//	resp.Schema = ProductTypeDataSourceSchema(ctx)
//}
//
//// Read refreshes the Terraform state with the latest data.
//func (d *productTypesDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
//	var state, plan ProductTypesDataSourceModel
//	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
//	if response.Diagnostics.HasError() {
//		return
//	}
//
//	params := ProductTypesFromTfToAPIReadParams(ctx, plan)
//	res := utils.ExecuteRequest(func() (*numspot.ReadProductTypesResponse, error) {
//		return numspotClient.ReadProductTypesWithResponse(ctx, d.provider.SpaceID, &params)
//	}, http.StatusOK, &response.Diagnostics)
//	if res == nil {
//		return
//	}
//	if res.JSON200.Items == nil {
//		response.Diagnostics.AddError("HTTP call failed", "got empty ProductTypes list")
//	}
//
//	objectItems, diags := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, ProductTypesFromHttpToTfDatasource)
//
//	if diags.HasError() {
//		response.Diagnostics.Append(diags...)
//		return
//	}
//
//	state = plan
//	state.Items = objectItems
//
//	response.Diagnostics.Append(response.State.Set(ctx, state)...)
//}
//
