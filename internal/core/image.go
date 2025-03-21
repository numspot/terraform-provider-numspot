package core

import (
	"context"
	"fmt"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

var (
	imagePendingStates = []string{creating, pending}
	imageTargetStates  = []string{available}
)

func CreateImage(ctx context.Context, provider *client.NumSpotSDK, body api.CreateImage, tags []api.ResourceTag, access *api.Access) (*api.Image, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *api.CreateImageResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, body, numspotClient.CreateImageWithResponse); err != nil {
		return nil, err
	}

	imageID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = createTags(ctx, provider, imageID, tags); err != nil {
			return nil, err
		}
	}

	if access != nil {
		if _, err = UpdateImageAccess(ctx, provider, imageID, *access); err != nil {
			return nil, err
		}
	}

	image, err := RetryReadImage(ctx, provider, imageID)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func UpdateImageAccess(ctx context.Context, provider *client.NumSpotSDK, id string, access api.Access) (*api.Image, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var body api.UpdateImageJSONRequestBody
	if *access.IsPublic { // If IsPublic is set to True
		body = api.UpdateImageJSONRequestBody{
			AccessCreation: api.AccessCreation{
				Additions: &api.Access{
					IsPublic: utils.EmptyTrueBoolPointer(),
				},
				Removals: nil,
			},
		}
	} else { // If IsPublic is set to False or removed
		body = api.UpdateImageJSONRequestBody{
			AccessCreation: api.AccessCreation{
				Additions: nil,
				Removals: &api.Access{
					IsPublic: utils.EmptyTrueBoolPointer(),
				},
			},
		}
	}

	updateImageResponse, err := numspotClient.UpdateImageWithResponse(ctx,
		provider.SpaceID,
		id,
		body,
	)
	if err != nil {
		return nil, err
	}

	return updateImageResponse.JSON200, nil
}

func UpdateImageTags(ctx context.Context, provider *client.NumSpotSDK, imageID string, stateTags []api.ResourceTag, planTags []api.ResourceTag) (*api.Image, error) {
	if err := updateResourceTags(ctx, provider, stateTags, planTags, imageID); err != nil {
		return nil, err
	}
	return ReadImageWithID(ctx, provider, imageID)
}

func DeleteImage(ctx context.Context, provider *client.NumSpotSDK, imageID string) (err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, imageID, numspotClient.DeleteImageWithResponse)
	if err != nil {
		return err
	}

	return nil
}

func ReadImageWithID(ctx context.Context, provider *client.NumSpotSDK, imageID string) (numSpotImage *api.Image, err error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	numSpotReadImage, err := numspotClient.ReadImagesByIdWithResponse(ctx, provider.SpaceID, imageID)
	if err != nil {
		return nil, err
	}
	if err = utils.ParseHTTPError(numSpotReadImage.Body, numSpotReadImage.StatusCode()); err != nil {
		return nil, err
	}

	return numSpotReadImage.JSON200, err
}

func RetryReadImage(ctx context.Context, provider *client.NumSpotSDK, imageID string) (*api.Image, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, imageID, provider.SpaceID, imagePendingStates, imageTargetStates, numspotClient.ReadImagesByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotImage, assert := read.(*api.Image)
	if !assert {
		return nil, fmt.Errorf("invalid image assertion %s", imageID)
	}
	return numSpotImage, err
}
