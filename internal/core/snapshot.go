package core

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	snapshotPendingStates = []string{pendingQueued, inQueue, pending}
	snapshotTargetStates  = []string{completed}

	// errInvalidSnapshotAssertion is returned when a snapshot read result cannot
	// be asserted to the expected type. The message intentionally mirrors the
	// historical wording.
	errInvalidSnapshotAssertion = errors.New("invalid nat gateway assertion")
	// errExpectedSnapshotList is returned when a snapshot list response is
	// unexpectedly nil.
	errExpectedSnapshotList = errors.New("HTTP call failed : expected a list of public Ips but got nil")
)

// CreateSnapshot creates snapshot.
func CreateSnapshot(
	ctx context.Context,
	provider *client.NumSpotSDK,
	body *api.CreateSnapshotJSONRequestBody,
	tags []api.ResourceTag,
) (numSpotSnapshot *api.Snapshot, err error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreate *api.CreateSnapshotResponse
	if retryCreate, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, *body, numspotClient.CreateSnapshotWithResponse); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	createdId := *retryCreate.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, createdId, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	return RetryReadSnapshot(ctx, provider, createdId)
}

// UpdateSnapshotTags updates snapshot tags.
func UpdateSnapshotTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	snapshotID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.Snapshot, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, snapshotID); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	return ReadSnapshot(ctx, provider, snapshotID)
}

// DeleteSnapshot deletes snapshot.
func DeleteSnapshot(ctx context.Context, provider *client.NumSpotSDK, snapshotID string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, snapshotID, numspotClient.DeleteSnapshotWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}
	return nil
}

// ReadSnapshot reads snapshot.
func ReadSnapshot(ctx context.Context, provider *client.NumSpotSDK, snapshotID string) (*api.Snapshot, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotSnapshot, err := numspotClient.ReadSnapshotsByIdWithResponse(ctx, provider.SpaceID, snapshotID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots by id")
	}

	if err = utils.ParseHTTPError(numSpotSnapshot.Body, numSpotSnapshot.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots by id")
	}

	return numSpotSnapshot.JSON200, nil
}

// RetryReadSnapshot retries read snapshot.
func RetryReadSnapshot(ctx context.Context, provider *client.NumSpotSDK, snapshotID string) (*api.Snapshot, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(ctx, snapshotID, provider.SpaceID, snapshotPendingStates, snapshotTargetStates,
		numspotClient.ReadSnapshotsByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numSpotSnapshot, assert := read.(*api.Snapshot)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidSnapshotAssertion, snapshotID)
	}
	return numSpotSnapshot, nil
}

// ReadSnapshotsWithParams reads snapshots with params.
func ReadSnapshotsWithParams(
	ctx context.Context,
	provider *client.NumSpotSDK,
	params *api.ReadSnapshotsParams,
) (numSpotSnapshot *[]api.Snapshot, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numSpotReadSnapshot, err := numspotClient.ReadSnapshotsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots")
	}
	if err = utils.ParseHTTPError(numSpotReadSnapshot.Body, numSpotReadSnapshot.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots")
	}
	if numSpotReadSnapshot.JSON200.Items == nil {
		return nil, errExpectedSnapshotList
	}

	return numSpotReadSnapshot.JSON200.Items, utils.WrapError(err, "failed to read snapshots")
}
