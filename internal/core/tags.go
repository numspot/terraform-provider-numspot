package core

import (
	"context"
	"net/http"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

// CreateTags Same as CreateTags but without using Diagnostics. Remove CreateTags when other function are reworked
func CreateTags(
	ctx context.Context,
	apiClient *numspot.ClientWithResponses,
	spaceId numspot.SpaceId,
	resourceId string,
	tags []numspot.ResourceTag,
) error {
	res, err := apiClient.CreateTagsWithResponse(ctx, spaceId, numspot.CreateTagsJSONRequestBody{
		ResourceIds: []string{resourceId},
		Tags:        tags,
	})
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusNoContent {
		apiError := utils.HandleError(res.Body)
		return apiError
	}

	return nil
}

func UpdateResourceTags(ctx context.Context, provider services.IProvider, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, resourceID string) (err error) {
	numSpotClient := provider.GetNumspotClient()
	spaceID := provider.GetSpaceID()

	toCreate, toDelete, toUpdate := Diff(stateTags, planTags)

	toDeleteApiTags := make([]numspot.ResourceTag, 0, len(toUpdate)+len(toDelete))
	toCreateApiTags := make([]numspot.ResourceTag, 0, len(toUpdate)+len(toCreate))
	for _, e := range toCreate {
		toCreateApiTags = append(toCreateApiTags, numspot.ResourceTag{
			Key:   e.Key,
			Value: e.Value,
		})
	}

	for _, e := range toDelete {
		toDeleteApiTags = append(toDeleteApiTags, numspot.ResourceTag{
			Key:   e.Key,
			Value: e.Value,
		})
	}

	for _, e := range toUpdate {
		// Delete
		toDeleteApiTags = append(toDeleteApiTags, numspot.ResourceTag{
			Key:   e.Key,
			Value: e.Value,
		})

		// Create
		toCreateApiTags = append(toCreateApiTags, numspot.ResourceTag{
			Key:   e.Key,
			Value: e.Value,
		})
	}

	if len(toDeleteApiTags) > 0 {
		if err = DeleteTags(
			ctx,
			numSpotClient,
			spaceID,
			resourceID,
			toDeleteApiTags,
		); err != nil {
			return err
		}
	}

	if len(toCreateApiTags) > 0 {
		if err = CreateTags(
			ctx,
			numSpotClient,
			spaceID,
			resourceID,
			toCreateApiTags,
		); err != nil {
			return err
		}
	}

	return nil
}

func DeleteTags(
	ctx context.Context,
	apiClient *numspot.ClientWithResponses,
	spaceId numspot.SpaceId,
	resourceId string,
	tags []numspot.ResourceTag,
) error {
	res, err := apiClient.DeleteTagsWithResponse(ctx, spaceId, numspot.DeleteTagsJSONRequestBody{
		ResourceIds: []string{resourceId},
		Tags:        tags,
	})
	if err != nil {
		return err
	}

	if res.StatusCode() != http.StatusNoContent {
		return utils.HandleError(res.Body)
	}

	return nil
}

// Diff calculates the differences between two slices of tags: which tags to create, delete, and update.
// Assumes that a tag's Key is unique in the slice.
func Diff(current, desired []numspot.ResourceTag) (toCreate, toDelete, toUpdate []numspot.ResourceTag) {
	currentMap := make(map[string]numspot.ResourceTag)
	desiredMap := make(map[string]numspot.ResourceTag)

	for _, tag := range current {
		currentMap[tag.Key] = tag
	}

	for _, tag := range desired {
		desiredMap[tag.Key] = tag
		if _, exists := currentMap[tag.Key]; !exists {
			toCreate = append(toCreate, tag)
		} else if currentMap[tag.Key].Value != tag.Value {
			toUpdate = append(toUpdate, tag)
		}
	}

	for _, tag := range current {
		if _, exists := desiredMap[tag.Key]; !exists {
			toDelete = append(toDelete, tag)
		}
	}

	return toCreate, toDelete, toUpdate
}
