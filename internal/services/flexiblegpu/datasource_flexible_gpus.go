package flexiblegpu

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-numspot/internal/core"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/services"
	"terraform-provider-numspot/internal/services/flexiblegpu/datasource_flexible_gpu"
	"terraform-provider-numspot/internal/utils"
)

var _ datasource.DataSource = &flexiblegpusDataSource{}

// flexiblegpusDataSource is the Terraform data source implementation for numspot_flexible_gpus.
type flexiblegpusDataSource struct {
	services.BaseDataSource
}

// NewFlexibleGpusDataSource returns a new flexiblegpusDataSource instance.
func NewFlexibleGpusDataSource() datasource.DataSource {
	return &flexiblegpusDataSource{}
}

// Metadata sets the data source type name.
func (d *flexiblegpusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flexible_gpus"
}

// Schema defines the schema for the data source.
func (d *flexiblegpusDataSource) Schema(ctx context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasource_flexible_gpu.FlexibleGpuDataSourceSchema(ctx)
}

// Read reads the data source items.
//
//nolint:gocritic // Terraform plugin framework interface requires value receiver
func (d *flexiblegpusDataSource) Read(ctx context.Context, request datasource.ReadRequest, response *datasource.ReadResponse) {
	var state, plan datasource_flexible_gpu.FlexibleGpuModel
	response.Diagnostics.Append(request.Config.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() {
		return
	}
	params := deserializeReadFlexibleGpus(ctx, &plan, &response.Diagnostics)
	if response.Diagnostics.HasError() {
		return
	}

	result, err := core.ReadFlexibleGpusWithParams(ctx, d.Provider, params)
	if err != nil {
		response.Diagnostics.AddError("unable to read flexible_gpus", err.Error())
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

func mappingItemsValue(ctx context.Context, item *api.FlexibleGpu) (datasource_flexible_gpu.ItemsValue, diag.Diagnostics) {
	value, valueDiags := datasource_flexible_gpu.NewItemsValue(datasource_flexible_gpu.ItemsValue{}.AttributeTypes(ctx), map[string]attr.Value{
		"availability_zone_name": types.StringValue(utils.ConvertAzNamePtrToString(item.AvailabilityZoneName)),
		"delete_on_vm_deletion":  types.BoolValue(*item.DeleteOnVmDeletion),
		"generation":             types.StringValue(utils.ConvertStringPtrToString(item.Generation)),
		"id":                     types.StringValue(utils.ConvertStringPtrToString(item.Id)),
		"model_name":             types.StringValue(utils.ConvertStringPtrToString(item.ModelName)),
		"state":                  types.StringValue(utils.ConvertStringPtrToString(item.State)),
		"vm_id":                  types.StringValue(utils.ConvertStringPtrToString(item.VmId)),
	})
	return value, valueDiags
}

func deserializeReadFlexibleGpus(
	ctx context.Context,
	tf *datasource_flexible_gpu.FlexibleGpuModel,
	diags *diag.Diagnostics,
) *api.ReadFlexibleGpusParams {
	return &api.ReadFlexibleGpusParams{
		DeleteOnVmDeletion:    tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generations:           utils.ConvertTfListToArrayOfString(ctx, tf.Generations, diags),
		ModelNames:            utils.ConvertTfListToArrayOfString(ctx, tf.ModelNames, diags),
		States:                utils.ConvertTfListToArrayOfString(ctx, tf.States, diags),
		VmIds:                 utils.ConvertTfListToArrayOfString(ctx, tf.VmIds, diags),
		AvailabilityZoneNames: utils.ConvertTfListToArrayOfAzName(ctx, tf.AvailabilityZoneNames, diags),
		Ids:                   utils.ConvertTfListToArrayOfString(ctx, tf.Ids, diags),
		Page: func() *api.PageParameter {
			if utils.IsTfValueNull(tf.Page) {
				return nil
			}
			return &api.PageParameter{NextToken: tf.Page.NextToken.ValueStringPointer(), Size: utils.FromTfInt64ToInt32Ptr(tf.Page.Size)}
		}(),
	}
}
