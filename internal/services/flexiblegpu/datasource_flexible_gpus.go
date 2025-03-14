package flexiblegpu

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &flexibleGpusDataSource{}
)

func (d *flexibleGpusDataSource) Configure(_ context.Context, request datasource.ConfigureRequest, response *datasource.ConfigureResponse) {
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

func NewFlexibleGpusDataSource() datasource.DataSource {
	return &flexibleGpusDataSource{}
}

type flexibleGpusDataSource struct {
	provider *client.NumSpotSDK
}

// Metadata returns the data source type name.
func (d *flexibleGpusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flexible_gpus"
}

// Schema defines the schema for the data source.
func (d *flexibleGpusDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = FlexibleGpuDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *flexibleGpusDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan FlexibleGpuDataSourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}

	numspotClient, err := d.provider.GetClient(ctx)
	if err != nil {
		response.Diagnostics.AddError("Error while initiating numspotClient", err.Error())
		return
	}
	params := deserializeFlexibleGPUDataSource(ctx, plan, &response.Diagnostics)

	res, err := numspotClient.ReadFlexibleGpusWithResponse(ctx, d.provider.SpaceID, &params)
	if err != nil {
		response.Diagnostics.AddError("unable to read flexible gpus", err.Error())
		return
	}

	if err = utils.ParseHTTPError(res.Body, res.StatusCode()); err != nil {
		response.Diagnostics.AddError("unable to read flexible gpus", err.Error())
		return
	}

	objectItems := utils.FromHttpGenericListToTfList(ctx, res.JSON200.Items, serializeFlexibleGPUDataSource, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeFlexibleGPUDataSource(ctx context.Context, tf FlexibleGpuDataSourceModel, diags *diag.Diagnostics) numspot.ReadFlexibleGpusParams {
	return numspot.ReadFlexibleGpusParams{
		States:                utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		Ids:                   utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.TfStringListToStringPtrList(ctx, tf.AvailabilityZoneNames, diags),
		DeleteOnVmDeletion:    utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
		Generations:           utils.TfStringListToStringPtrList(ctx, tf.Generations, diags),
		ModelNames:            utils.TfStringListToStringPtrList(ctx, tf.ModelNames, diags),
		VmIds:                 utils.TfStringListToStringPtrList(ctx, tf.VmIds, diags),
	}
}

func serializeFlexibleGPUDataSource(_ context.Context, http *numspot.FlexibleGpu, _ *diag.Diagnostics) *FlexibleGpuModelItemDataSource {
	return &FlexibleGpuModelItemDataSource{
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		Id:                   types.StringPointerValue(http.Id),
		State:                types.StringPointerValue(http.State),
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		ModelName:            types.StringPointerValue(http.ModelName),
		VmId:                 types.StringPointerValue(http.VmId),
	}
}
