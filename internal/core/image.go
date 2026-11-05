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
	imagePendingStates = []string{creating, pending}
	imageTargetStates  = []string{available}

	// errInvalidImageAssertion is returned when an image read result cannot be
	// asserted to the expected type.
	errInvalidImageAssertion = errors.New("invalid image assertion")
)

// CreateImage creates image.
func CreateImage(ctx context.Context, provider *client.NumspotSDK, body *api.CreateImage, tags []api.ResourceTag) (*api.Image, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	var retryCreateResponse *api.CreateImageResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		spaceID,
		*body,
		numspotClient.CreateImageWithResponse,
	); err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	imageID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, imageID, tags); err != nil {
			return nil, utils.WrapError(err, "failed to get API client")
		}
	}

	image, err := RetryReadImage(ctx, provider, imageID)
	if err != nil {
		return nil, err
	}

	return image, nil
}

// UpdateImageTags updates image tags.
func UpdateImageTags(
	ctx context.Context,
	provider *client.NumspotSDK,
	imageID string,
	stateTags []api.ResourceTag,
	planTags []api.ResourceTag,
) (*api.Image, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, imageID); err != nil {
		return nil, err
	}
	return ReadImage(ctx, provider, imageID)
}

// DeleteImage deletes image.
func DeleteImage(ctx context.Context, provider *client.NumspotSDK, imageID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, imageID, numspotClient.DeleteImageWithResponse)
	if err != nil {
		return utils.WrapError(err, "failed to delete resource")
	}

	return nil
}

// ReadImage reads image.
func ReadImage(ctx context.Context, provider *client.NumspotSDK, imageID string) (numspotImage *api.Image, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	numspotReadImage, err := numspotClient.ReadImagesByIdWithResponse(ctx, provider.SpaceID, imageID)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read images by id")
	}
	if err = utils.ParseHTTPError(numspotReadImage.Body, numspotReadImage.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read images by id")
	}

	return numspotReadImage.JSON200, utils.WrapError(err, "failed to read images by id")
}

// RetryReadImage retries read image.
func RetryReadImage(ctx context.Context, provider *client.NumspotSDK, imageID string) (*api.Image, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}
	read, err := utils.RetryReadUntilStateValid(
		ctx,
		imageID,
		provider.SpaceID,
		imagePendingStates,
		imageTargetStates,
		numspotClient.ReadImagesByIdWithResponse,
	)
	if err != nil {
		return nil, utils.WrapError(err, "failed to wait for resource state")
	}

	numspotImage, assert := read.(*api.Image)
	if !assert {
		return nil, fmt.Errorf("%w %s", errInvalidImageAssertion, imageID)
	}
	return numspotImage, utils.WrapError(err, "failed to wait for resource state")
}
