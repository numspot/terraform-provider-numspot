package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/datasource_snapshot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_snapshot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SnapshotFromHttpToTf(http *iaas.Snapshot) resource_snapshot.SnapshotModel {
	var creationDateStr *string

	if http.CreationDate != nil {
		tmp := (*http.CreationDate).String()
		creationDateStr = &tmp
	}

	return resource_snapshot.SnapshotModel{
		CreationDate: types.StringPointerValue(creationDateStr),
		Description:  types.StringPointerValue(http.Description),
		Id:           types.StringPointerValue(http.Id),
		Progress:     utils.FromIntPtrToTfInt64(http.Progress),
		State:        types.StringPointerValue(http.State),
		VolumeId:     types.StringPointerValue(http.VolumeId),
		VolumeSize:   utils.FromIntPtrToTfInt64(http.VolumeSize),
	}
}

func SnapshotFromTfToCreateRequest(tf *resource_snapshot.SnapshotModel) iaas.CreateSnapshotJSONRequestBody {
	return iaas.CreateSnapshotJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		SourceRegionName: tf.SourceRegionName.ValueStringPointer(),
		SourceSnapshotId: tf.SourceSnapshotId.ValueStringPointer(),
		VolumeId:         tf.VolumeId.ValueStringPointer(),
	}
}

func SnapshotsFromTfToAPIReadParams(ctx context.Context, tf SnapshotsDataSourceModel) iaas.ReadSnapshotsParams {
	return iaas.ReadSnapshotsParams{
		Descriptions:     utils.TfStringListToStringPtrList(ctx, tf.Descriptions),
		FromCreationDate: utils.FromTfStringToStringPtr(tf.FromCreationDate),
		PermissionsToCreateVolumeGlobalPermission: utils.FromTfBoolToBoolPtr(tf.PermissionsToCreateVolumeGlobalPermission),
		Progresses:     utils.TFInt64ListToIntListPointer(ctx, tf.Progresses),
		States:         utils.TfStringListToStringPtrList(ctx, tf.States),
		ToCreationDate: utils.FromTfStringToStringPtr(tf.ToCreationDate),
		VolumeIds:      utils.TfStringListToStringPtrList(ctx, tf.VolumeIds),
		VolumeSizes:    utils.TFInt64ListToIntListPointer(ctx, tf.VolumeSizes),
		TagKeys:        utils.TfStringListToStringPtrList(ctx, tf.TagKeys),
		TagValues:      utils.TfStringListToStringPtrList(ctx, tf.TagValues),
		Tags:           utils.TfStringListToStringPtrList(ctx, tf.Tags),
		Ids:            utils.TfStringListToStringPtrList(ctx, tf.IDs),
	}
}

func SnapshotsFromHttpToTfDatasource(ctx context.Context, http *iaas.Snapshot) (*datasource_snapshot.SnapshotModel, diag.Diagnostics) {
	var (
		diags          diag.Diagnostics
		tagsList       types.List
		creationDateTf types.String
		progressTf     types.Int64
		volumeSizeTf   types.Int64
	)
	if http.Tags != nil {
		tagsList, diags = utils.GenericListToTfListValue(ctx, tags.TagsValue{}, tags.ResourceTagFromAPI, *http.Tags)
		if diags.HasError() {
			return nil, diags
		}
	}

	// Creation Date
	if http.CreationDate != nil {
		date := *http.CreationDate
		creationDateTf = types.StringValue(date.Format(time.RFC3339))
	}

	if http.Progress != nil {
		progress := int64(*http.Progress)
		progressTf = types.Int64PointerValue(&progress)
	}

	if http.VolumeSize != nil {
		volumeSize := int64(*http.VolumeSize)
		volumeSizeTf = types.Int64PointerValue(&volumeSize)
	}

	return &datasource_snapshot.SnapshotModel{
		Id:           types.StringPointerValue(http.Id),
		State:        types.StringPointerValue(http.State),
		Tags:         tagsList,
		CreationDate: creationDateTf,
		Description:  types.StringPointerValue(http.Description),
		VolumeId:     types.StringPointerValue(http.VolumeId),
		Progress:     progressTf,
		VolumeSize:   volumeSizeTf,
	}, nil
}
