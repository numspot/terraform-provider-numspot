package snapshot

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services/tags"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SnapshotFromHttpToTf(ctx context.Context, http *numspot.Snapshot, diags *diag.Diagnostics) *SnapshotModel {
	var (
		tagsTf          types.List
		creationDateStr *string
	)

	if http.CreationDate != nil {
		tmp := (*http.CreationDate).String()
		creationDateStr = &tmp
	}

	if http.Tags != nil {
		tagsTf = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
		}
	}

	return &SnapshotModel{
		CreationDate: types.StringPointerValue(creationDateStr),
		Description:  types.StringPointerValue(http.Description),
		Id:           types.StringPointerValue(http.Id),
		Progress:     utils.FromIntPtrToTfInt64(http.Progress),
		State:        types.StringPointerValue(http.State),
		VolumeId:     types.StringPointerValue(http.VolumeId),
		VolumeSize:   utils.FromIntPtrToTfInt64(http.VolumeSize),
		Tags:         tagsTf,
	}
}

func SnapshotFromTfToCreateRequest(tf *SnapshotModel) numspot.CreateSnapshotJSONRequestBody {
	return numspot.CreateSnapshotJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		SourceRegionName: tf.SourceRegionName.ValueStringPointer(),
		SourceSnapshotId: tf.SourceSnapshotId.ValueStringPointer(),
		VolumeId:         tf.VolumeId.ValueStringPointer(),
	}
}

func SnapshotsFromTfToAPIReadParams(ctx context.Context, tf SnapshotsDataSourceModel, diags *diag.Diagnostics) numspot.ReadSnapshotsParams {
	return numspot.ReadSnapshotsParams{
		Descriptions:     utils.TfStringListToStringPtrList(ctx, tf.Descriptions, diags),
		FromCreationDate: utils.FromTfStringToStringPtr(tf.FromCreationDate),
		IsPublic:         utils.FromTfBoolToBoolPtr(tf.IsPublic),
		Progresses:       utils.TFInt64ListToIntListPointer(ctx, tf.Progresses, diags),
		States:           utils.TfStringListToStringPtrList(ctx, tf.States, diags),
		ToCreationDate:   utils.FromTfStringToStringPtr(tf.ToCreationDate),
		VolumeIds:        utils.TfStringListToStringPtrList(ctx, tf.VolumeIds, diags),
		VolumeSizes:      utils.TFInt64ListToIntListPointer(ctx, tf.VolumeSizes, diags),
		TagKeys:          utils.TfStringListToStringPtrList(ctx, tf.TagKeys, diags),
		TagValues:        utils.TfStringListToStringPtrList(ctx, tf.TagValues, diags),
		Tags:             utils.TfStringListToStringPtrList(ctx, tf.Tags, diags),
		Ids:              utils.TfStringListToStringPtrList(ctx, tf.Ids, diags),
	}
}

func SnapshotsFromHttpToTfDatasource(ctx context.Context, http *numspot.Snapshot, diags *diag.Diagnostics) *SnapshotModelDatasource {
	var (
		tagsList       types.List
		creationDateTf types.String
		progressTf     types.Int64
		volumeSizeTf   types.Int64
	)

	if http.Tags != nil {
		tagsList = utils.GenericListToTfListValue(ctx, tags.ResourceTagFromAPI, *http.Tags, diags)
		if diags.HasError() {
			return nil
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

	access, err := NewAccessValue(AccessValue{}.AttributeTypes(ctx),
		map[string]attr.Value{
			"is_public": types.BoolPointerValue(http.Access.IsPublic),
		})
	if err != nil {
		return nil
	}

	return &SnapshotModelDatasource{
		Id:           types.StringPointerValue(http.Id),
		State:        types.StringPointerValue(http.State),
		Tags:         tagsList,
		CreationDate: creationDateTf,
		Description:  types.StringPointerValue(http.Description),
		VolumeId:     types.StringPointerValue(http.VolumeId),
		Progress:     progressTf,
		VolumeSize:   volumeSizeTf,
		Access:       access,
	}
}
