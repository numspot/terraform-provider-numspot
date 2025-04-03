package flexiblegpu

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services/flexiblegpu/datasource_flexible_gpu"
	"terraform-provider-numspot/internal/utils"
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
			"Unexpected Datasource Configure Type",
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
	resp.Schema = datasource_flexible_gpu.FlexibleGpuDataSourceSchema(ctx)
}

// Read refreshes the Terraform state with the latest data.
func (d *flexibleGpusDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_flexible_gpu.FlexibleGpuModel
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

	objectItems, serializeDiags := utils.SerializeDatasourceItems(ctx, *res.JSON200.Items, mappingItemsValue)
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

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeFlexibleGPUDataSource(ctx context.Context, tf datasource_flexible_gpu.FlexibleGpuModel, diags *diag.Diagnostics) api.ReadFlexibleGpusParams {
	return api.ReadFlexibleGpusParams{
		States:                utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		Ids:                   utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		DeleteOnVmDeletion:    tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generations:           utils.ConvertTfListToArrayOfString(ctx, tf.Generations, diags),
		ModelNames:            utils.ConvertTfListToArrayOfString(ctx, tf.ModelNames, diags),
		VmIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.VmIds, diags),
	}
}

func mappingItemsValue(ctx context.Context, flexiblGpus api.FlexibleGpu) (datasource_flexible_gpu.ItemsValue, diag.Diagnostics) {
	return datasource_flexible_gpu.NewItemsValue(datasource_flexible_gpu.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(flexiblGpus.AvailabilityZoneName)),
		"delete_on_vm_deletion":  types.BoolPointerValue(flexiblGpus.DeleteOnVmDeletion),
		"generation":             types.StringValue(utils.ConvertStringPtrToString(flexiblGpus.Generation)),
		"id":                     types.StringValue(utils.ConvertStringPtrToString(flexiblGpus.Id)),
		"model_name":             types.StringValue(utils.ConvertStringPtrToString(flexiblGpus.ModelName)),
		"state":                  types.StringValue(utils.ConvertStringPtrToString(flexiblGpus.State)),
		"vm_id":                  types.StringValue(utils.ConvertStringPtrToString(flexiblGpus.VmId)),
	})
}
