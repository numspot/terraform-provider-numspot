package provider

import (
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/conns/api"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider/resource_snapshot"
)

func SnapshotFromTfToHttp(tf *resource_snapshot.SnapshotModel) *api.SnapshotSchema {
	return &api.SnapshotSchema{}
}

func SnapshotFromHttpToTf(http *api.SnapshotSchema) resource_snapshot.SnapshotModel {
	return resource_snapshot.SnapshotModel{}
}

func SnapshotFromTfToCreateRequest(tf *resource_snapshot.SnapshotModel) api.CreateSnapshotJSONRequestBody {
	return api.CreateSnapshotJSONRequestBody{}
}
