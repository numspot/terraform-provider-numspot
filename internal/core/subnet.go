package core

import (
	"context"
	"errors"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/client"
	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateSubnet(ctx context.Context, provider *client.NumSpotSDK, payload numspot.CreateSubnet, mapPublicIPOnLaunch bool, tags []numspot.ResourceTag) (*numspot.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := utils.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.SpaceID,
		payload,
		numspotClient.CreateSubnetWithResponse)
	if err != nil {
		return nil, err
	}

	createdID := *res.JSON201.Id

	if mapPublicIPOnLaunch {
		if _, err = UpdateSubnetAttributes(ctx, provider, createdID, mapPublicIPOnLaunch); err != nil {
			return nil, fmt.Errorf("failed to update MapPublicIPOnLaunch: %w", err)
		}
	}

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, createdID, tags); err != nil {
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	resRead, err := RetryReadSubnet(ctx, provider, createdID)
	if err != nil {
		return nil, fmt.Errorf("failed to read subnet: %w", err)
	}

	return resRead, nil
}

func RetryReadSubnet(ctx context.Context, provider *client.NumSpotSDK, id string) (*numspot.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := utils.RetryReadUntilStateValid(
		ctx,
		id,
		provider.SpaceID,
		[]string{pending},
		[]string{available},
		numspotClient.ReadSubnetsByIdWithResponse,
	)
	if err != nil {
		return nil, err
	}

	res, ok := read.(*numspot.Subnet)
	if !ok {
		return nil, errors.New("object conversion error")
	}

	return res, nil
}

func ReadSubnet(ctx context.Context, provider *client.NumSpotSDK, id string) (*numspot.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.ReadSubnetsByIdWithResponse(ctx, provider.SpaceID, id)
	if err != nil {
		return nil, err
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf("got %s", res.Status())
	}

	return res.JSON200, nil
}

func UpdateSubnetAttributes(ctx context.Context, provider *client.NumSpotSDK, id string, mapPublicIpOnLaunch bool) (*numspot.Subnet, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := numspotClient.UpdateSubnetWithResponse(ctx, provider.SpaceID, id, numspot.UpdateSubnetJSONRequestBody{
		MapPublicIpOnLaunch: mapPublicIpOnLaunch,
	})
	if err != nil {
		return nil, err
	}
	if res.JSON200 == nil {
		return nil, fmt.Errorf("got %s", res.Status())
	}

	resRead, err := ReadSubnet(ctx, provider, *res.JSON200.Id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read subnet %w", err)
	}

	return resRead, nil
}

func UpdateSubnetTags(ctx context.Context, provider *client.NumSpotSDK, id string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.Subnet, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, id); err != nil {
		return nil, err
	}
	return ReadSubnet(ctx, provider, id)
}

func DeleteSubnet(ctx context.Context, provider *client.NumSpotSDK, id string) error {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	err = utils.RetryDeleteUntilResourceAvailable(ctx, provider.SpaceID, id, numspotClient.DeleteSubnetWithResponse)
	if err != nil {
		return err
	}

	return nil
}
