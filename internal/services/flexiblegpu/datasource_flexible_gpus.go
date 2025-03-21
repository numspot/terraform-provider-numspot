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

	objectItems := serializeFlexibleGPUDataSource(ctx, res.JSON200.Items, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	state = plan
	state.Items = objectItems.Items

	response.Diagnostics.Append(response.State.Set(ctx, state)...)
}

func deserializeFlexibleGPUDataSource(ctx context.Context, tf datasource_flexible_gpu.FlexibleGpuModel, diags *diag.Diagnostics) api.ReadFlexibleGpusParams {
	return api.ReadFlexibleGpusParams{
		States:                utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		Ids:                   utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		AvailabilityZoneNames: utils.ConvertTfListToArrayOfString(ctx, tf.AvailabilityZoneNames, diags),
		DeleteOnVmDeletion:    tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generations:           utils.ConvertTfListToArrayOfString(ctx, tf.Generations, diags),
		ModelNames:            utils.ConvertTfListToArrayOfString(ctx, tf.ModelNames, diags),
		VmIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.VmIds, diags),
	}
}

func serializeFlexibleGPUDataSource(ctx context.Context, flexiblGpus *[]api.FlexibleGpu, diags *diag.Diagnostics) datasource_flexible_gpu.FlexibleGpuModel {
	var flexibleGpusList types.List
	var serializeDiags diag.Diagnostics

	if len(*flexiblGpus) != 0 {
		ll := len(*flexiblGpus)
		itemsValue := make([]datasource_flexible_gpu.ItemsValue, ll)

		for i := 0; ll > i; i++ {
			itemsValue[i], serializeDiags = datasource_flexible_gpu.NewItemsValue(datasource_flexible_gpu.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
				"availability_zone_name": types.StringValue(utils.ConvertStringPtrToString((*flexiblGpus)[i].AvailabilityZoneName)),
				"delete_on_vm_deletion":  types.BoolPointerValue((*flexiblGpus)[i].DeleteOnVmDeletion),
				"generation":             types.StringValue(utils.ConvertStringPtrToString((*flexiblGpus)[i].Generation)),
				"id":                     types.StringValue(utils.ConvertStringPtrToString((*flexiblGpus)[i].Id)),
				"model_name":             types.StringValue(utils.ConvertStringPtrToString((*flexiblGpus)[i].ModelName)),
				"state":                  types.StringValue(utils.ConvertStringPtrToString((*flexiblGpus)[i].State)),
				"vm_id":                  types.StringValue(utils.ConvertStringPtrToString((*flexiblGpus)[i].VmId)),
			})
			if serializeDiags.HasError() {
				diags.Append(serializeDiags...)
				continue
			}
		}

		flexibleGpusList, serializeDiags = types.ListValueFrom(ctx, new(datasource_flexible_gpu.ItemsValue).Type(ctx), itemsValue)
		if serializeDiags.HasError() {
			diags.Append(serializeDiags...)
		}
	} else {
		flexibleGpusList = types.ListNull(new(datasource_flexible_gpu.ItemsValue).Type(ctx))
	}

	return datasource_flexible_gpu.FlexibleGpuModel{
		Items: flexibleGpusList,
	}
}
