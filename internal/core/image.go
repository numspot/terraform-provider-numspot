package core

import (
	"context"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

var (
	imagePendingStates = []string{creating, pending}
	imageTargetStates  = []string{available}
)

func CreateImage(ctx context.Context, provider *client.NumSpotSDK, body numspot.CreateImage, tags []numspot.ResourceTag, access *numspot.Access) (*numspot.Image, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var retryCreateResponse *numspot.CreateImageResponse
	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, body, numspotClient.CreateImageWithResponse); err != nil {
		return nil, err
	}

	imageID := *retryCreateResponse.JSON201.Id

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, imageID, tags); err != nil {
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

func UpdateImageAccess(ctx context.Context, provider *client.NumSpotSDK, id string, access numspot.Access) (*numspot.Image, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	var body numspot.UpdateImageJSONRequestBody
	if *access.IsPublic { // If IsPublic is set to True
		body = numspot.UpdateImageJSONRequestBody{
			AccessCreation: numspot.AccessCreation{
				Additions: &numspot.Access{
					IsPublic: utils.EmptyTrueBoolPointer(),
				},
				Removals: nil,
			},
		}
	} else { // If IsPublic is set to False or removed
		body = numspot.UpdateImageJSONRequestBody{
			AccessCreation: numspot.AccessCreation{
				Additions: nil,
				Removals: &numspot.Access{
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

func UpdateImageTags(ctx context.Context, provider *client.NumSpotSDK, imageID string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.Image, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, imageID); err != nil {
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

func ReadImageWithID(ctx context.Context, provider *client.NumSpotSDK, imageID string) (numSpotImage *numspot.Image, err error) {
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

func RetryReadImage(ctx context.Context, provider *client.NumSpotSDK, imageID string) (*numspot.Image, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}
	read, err := utils.RetryReadUntilStateValid(ctx, imageID, provider.SpaceID, imagePendingStates, imageTargetStates, numspotClient.ReadImagesByIdWithResponse)
	if err != nil {
		return nil, err
	}

	numSpotImage, assert := read.(*numspot.Image)
	if !assert {
		return nil, fmt.Errorf("invalid image assertion %s", imageID)
	}
	return numSpotImage, err
}
