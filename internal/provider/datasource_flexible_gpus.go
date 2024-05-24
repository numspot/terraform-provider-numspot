package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_flexible_gpu"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

type FlexibleGpuModel struct {
	Items                 []datasource_flexible_gpu.FlexibleGpuModel `tfsdk:"items"`
	AvailabilityZoneNames types.List                                 `tfsdk:"availability_zone_names"`
	DeleteOnVmDeletion    types.Bool                                 `tfsdk:"delete_on_vm_deletion"`
	Generations           types.List                                 `tfsdk:"generations"`
	Ids                   types.List                                 `tfsdk:"ids"`
	ModelNames            types.List                                 `tfsdk:"model_names"`
	States                types.List                                 `tfsdk:"states"`
	VmIds                 types.List                                 `tfsdk:"vm_ids"`
}

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &flexibleGpusDataSource{}
)

func (d *flexibleGpusDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewFlexibleGpusDataSource() datasource.DataSource {
	return &flexibleGpusDataSource{}
}

type flexibleGpusDataSource struct {
	provider Provider
}

// Metadata returns the data source type name.
func (d *flexibleGpusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flexible_gpus"
}

// Schema defines the schema for the data source.
func (d *flexibleGpusDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_flexible_gpu.FlexibleGpuDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *flexibleGpusDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan FlexibleGpuModel
	request.Config.Get(ctx, &plan)

	params := FlexibleGpusFromTfToAPIReadParams(ctx, plan)
	res := utils.ExecuteRequest(func() (*iaas.ReadFlexibleGpusResponse, error) {
		return d.provider.ApiClient.ReadFlexibleGpusWithResponse(ctx, d.provider.SpaceID, &params)
	}, http.StatusOK, &response.Diagnostics)
	if res == nil {
		return
	}
	if res.JSON200.Items == nil {
		response.Diagnostics.AddError("HTTP call failed", "got empty FlexibleGpus list")
	}

	for _, item := range *res.JSON200.Items {
		tf, diags := FlexibleGpusFromHttpToTfDatasource(ctx, &item)
		if diags != nil {
			response.Diagnostics.AddError("Error while converting FlexibleGpu HTTP object to Terraform object", diags.Errors()[0].Detail())
		}
		state.Items = append(state.Items, *tf)
	}

	state = plan

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}
