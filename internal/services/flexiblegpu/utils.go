package flexiblegpu

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func FlexibleGpuFromHttpToTf(http *numspot.FlexibleGpu) FlexibleGpuModel {
	return FlexibleGpuModel{
		DeleteOnVmDeletion:   types.BoolPointerValue(http.DeleteOnVmDeletion),
		Generation:           types.StringPointerValue(http.Generation),
		Id:                   types.StringPointerValue(http.Id),
		ModelName:            types.StringPointerValue(http.ModelName),
		State:                types.StringPointerValue(http.State),
		AvailabilityZoneName: types.StringPointerValue(http.AvailabilityZoneName),
		VmId:                 types.StringPointerValue(http.VmId),
	}
}

func FlexibleGpuFromTfToCreateRequest(tf *FlexibleGpuModel) numspot.CreateFlexibleGpuJSONRequestBody {
	return numspot.CreateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion:   tf.DeleteOnVmDeletion.ValueBoolPointer(),
		Generation:           tf.Generation.ValueStringPointer(),
		ModelName:            tf.ModelName.ValueString(),
		AvailabilityZoneName: tf.AvailabilityZoneName.ValueString(),
	}
}

func FlexibleGpuFromTfToUpdateRequest(tf *FlexibleGpuModel) numspot.UpdateFlexibleGpuJSONRequestBody {
	return numspot.UpdateFlexibleGpuJSONRequestBody{
		DeleteOnVmDeletion: utils.FromTfBoolToBoolPtr(tf.DeleteOnVmDeletion),
	}
}

func FlexibleGpusFromTfToAPIReadParams(ctx context.Context, tf FlexibleGpuDataSourceModel, diags *diag.Diagnostics) numspot.ReadFlexibleGpusParams {
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

func FlexibleGpusFromHttpToTfDatasource(ctx context.Context, http *numspot.FlexibleGpu, diags *diag.Diagnostics) *FlexibleGpuModelItemDataSource {
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

func LinkFlexibleGpuFromTfToCreateRequest(tf *FlexibleGpuModel) numspot.LinkFlexibleGpuJSONRequestBody {
	vmId := utils.FromTfStringToStringPtr(tf.VmId)
	return numspot.LinkFlexibleGpuJSONRequestBody{
		VmId: utils.GetPtrValue(vmId),
	}
}
