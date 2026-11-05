package core

import (
	"context"
	"errors"
	"fmt"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
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
	provider *client.NumspotSDK,
	body *api.CreateSnapshotJSONRequestBody,
	tags []api.ResourceTag,
) (numspotSnapshot *api.Snapshot, err error) {
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
	provider *client.NumspotSDK,
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
func DeleteSnapshot(ctx context.Context, provider *client.NumspotSDK, snapshotID string) error {
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
func ReadSnapshot(ctx context.Context, provider *client.NumspotSDK, snapshotID string) (*api.Snapshot, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotSnapshot, err := numspotClient.ReadSnapshotsByIdWithResponse(ctx, provider.SpaceID, snapshotID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots by id")
	}

	if err = utils.ParseHTTPError(numspotSnapshot.Body, numspotSnapshot.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots by id")
	}

	return numspotSnapshot.JSON200, nil
}

// RetryReadSnapshot retries read snapshot.
func RetryReadSnapshot(ctx context.Context, provider *client.NumspotSDK, snapshotID string) (*api.Snapshot, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(ctx, snapshotID, provider.SpaceID, snapshotPendingStates, snapshotTargetStates,
		numspotClient.ReadSnapshotsByIdWithResponse)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numspotSnapshot, assert := read.(*api.Snapshot)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidSnapshotAssertion, snapshotID)
	}
	return numspotSnapshot, nil
}

// ReadSnapshotsWithParams reads snapshots with params.
func ReadSnapshotsWithParams(
	ctx context.Context,
	provider *client.NumspotSDK,
	params *api.ReadSnapshotsParams,
) (numspotSnapshot *[]api.Snapshot, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadSnapshot, err := numspotClient.ReadSnapshotsWithResponse(ctx, provider.SpaceID, params)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots")
	}
	if err = utils.ParseHTTPError(numspotReadSnapshot.Body, numspotReadSnapshot.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read snapshots")
	}
	if numspotReadSnapshot.JSON200.Items == nil {
		return nil, errExpectedSnapshotList
	}

	return numspotReadSnapshot.JSON200.Items, utils.WrapError(err, "failed to read snapshots")
}
