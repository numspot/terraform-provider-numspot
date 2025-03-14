package core

import (
	"context"
	"fmt"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

var (
	snapshotPendingStates = []string{pendingQueued, inQueue, pending}
	snapshotTargetStates  = []string{completed}
)

func CreateSnapshot(ctx context.Context, provider *client.NumSpotSDK, tags []numspot.ResourceTag, body numspot.CreateSnapshotJSONRequestBody) (numSpotSnapshot *numspot.Snapshot, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreate *numspot.CreateSnapshotResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, body, numspotClient.CreateSnapshotWithResponse); err != nil {
		return nil, err
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, err
		}
	}

	return RetryReadSnapshot(ctx, provider, createdId)
}

func UpdateSnapshotTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, snapshotID string) (*numspot.Snapshot, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, snapshotID); err != nil {
		return nil, err
	}
	return ReadSnapshot(ctx, provider, snapshotID)
}

func DeleteSnapshot(ctx context.Context, provider *client.NumSpotSDK, snapshotID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, snapshotID, numspotClient.DeleteSnapshotWithResponse)
	if err != nil {
		return err
	}
	return nil
}

func ReadSnapshot(ctx context.Context, provider *client.NumSpotSDK, snapshotID string) (*numspot.Snapshot, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotSnapshot, err := numspotClient.ReadSnapshotsByIdWithResponse(ctx, provider.SpaceID, snapshotID)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(numSpotSnapshot.Body, numSpotSnapshot.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotSnapshot.JSON200, nil
}

func RetryReadSnapshot(ctx context.Context, provider *client.NumSpotSDK, snapshotID string) (*numspot.Snapshot, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, snapshotID, provider.SpaceID, snapshotPendingStates, snapshotTargetStates,
		numspotClient.ReadSnapshotsByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotSnapshot, assert := read.(*numspot.Snapshot)
	if !assert {
		return nil, fmt.Errorf("invalid nat gateway assertion %s", snapshotID)
	}
	return numSpotSnapshot, nil
}

func ReadSnapshotsWithParams(ctx context.Context, provider *client.NumSpotSDK, params numspot.ReadSnapshotsParams) (numSpotSnapshot *[]numspot.Snapshot, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadSnapshot, err := numspotClient.ReadSnapshotsWithResponse(ctx, provider.SpaceID, &params)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadSnapshot.Body, numSpotReadSnapshot.StatusCode()); err != nil {
		return nil, err
	}
	if numSpotReadSnapshot.JSON200.Items == nil {
		return nil, fmt.Errorf("HTTP call failed : expected a list of public Ips but got nil")
	}

	return numSpotReadSnapshot.JSON200.Items, err
}
