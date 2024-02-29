package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_snapshot"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func SnapshotFromTfToHttp(tf *resource_snapshot.SnapshotModel) *api.Snapshot {
	return &api.Snapshot{}
}

func SnapshotFromHttpToTf(http *api.Snapshot) resource_snapshot.SnapshotModel {
	var creationDateStr *string

	if http.CreationDate != nil {
		tmp := (*http.CreationDate).String()
		creationDateStr = &tmp
	}

	return resource_snapshot.SnapshotModel{
		AccountAlias: types.StringPointerValue(http.AccountAlias),
		AccountId:    types.StringPointerValue(http.AccountId),
		CreationDate: types.StringPointerValue(creationDateStr),
		Description:  types.StringPointerValue(http.Description),
		Id:           types.StringPointerValue(http.Id),
		// PermissionsToCreateVolume: resource_snapshot.PermissionsToCreateVolumeValue{}, ??
		Progress:   utils.FromIntPtrToTfInt64(http.Progress),
		State:      types.StringPointerValue(http.State),
		VolumeId:   types.StringPointerValue(http.VolumeId),
		VolumeSize: utils.FromIntPtrToTfInt64(http.VolumeSize),
		//SourceRegionName: must be set from creation body
		//SourceSnapshotId: must be set from creation body
	}
}

func SnapshotFromTfToCreateRequest(tf *resource_snapshot.SnapshotModel) api.CreateSnapshotJSONRequestBody {
	return api.CreateSnapshotJSONRequestBody{
		Description:      tf.Description.ValueStringPointer(),
		SourceRegionName: tf.SourceRegionName.ValueStringPointer(),
		SourceSnapshotId: tf.SourceSnapshotId.ValueStringPointer(),
		VolumeId:         tf.VolumeId.ValueStringPointer(),
	}
}
