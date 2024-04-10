package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/iaas"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_snapshot"
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
