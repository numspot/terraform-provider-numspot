package core

import (
	"context"
	"net/http"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud-sdk/numspot-sdk-go/pkg/numspot"

	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/client"
	"gitlab.tooling.cloudgouv-eu-west-1.numspot.internal/cloud/terraform-provider-numspot/internal/utils"
)

// CreateTags Same as CreateTags but without using Diagnostics. Remove CreateTags when other function are reworked
func createTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	resourceId string,
	tags []numspot.ResourceTag,
) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	res, err := numspotClient.CreateTagsWithResponse(ctx, provider.SpaceID, numspot.CreateTagsJSONRequestBody{
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

func updateResourceTags(ctx context.Context, provider *client.NumSpotSDK, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag, resourceID string) (err error) {
	toCreate, toDelete, toUpdate := diff(stateTags, planTags)

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
		if err = deleteTags(
			ctx,
			provider,
			resourceID,
			toDeleteApiTags,
		); err != nil {
			return err
		}
	}

	if len(toCreateApiTags) > 0 {
		if err = createTags(
			ctx,
			provider,
			resourceID,
			toCreateApiTags,
		); err != nil {
			return err
		}
	}

	return nil
}

func deleteTags(
	ctx context.Context,
	provider *client.NumSpotSDK,
	resourceId string,
	tags []numspot.ResourceTag,
) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}
	res, err := numspotClient.DeleteTagsWithResponse(ctx, provider.SpaceID, numspot.DeleteTagsJSONRequestBody{
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
func diff(current, desired []numspot.ResourceTag) (toCreate, toDelete, toUpdate []numspot.ResourceTag) {
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
