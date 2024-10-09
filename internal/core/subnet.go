package core

import (
	"context"
	"errors"
	"fmt"

	"gitlab.numspot.cloud/cloud/numspot-sdk-go/pkg/numspot"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/services"
	utils2 "gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/utils"
)

func CreateSubnet(ctx context.Context, provider services.IProvider, payload numspot.CreateSubnet, mapPublicIPOnLaunch bool, tags []numspot.ResourceTag) (*numspot.Subnet, error) {
	res, err := utils2.RetryCreateUntilResourceAvailableWithBody(
		ctx,
		provider.GetSpaceID(),
		payload,
		provider.GetNumspotClient().CreateSubnetWithResponse)
	if err != nil {
		return nil, err
	}

	createdID := *res.JSON201.Id
	_, err = RetryReadSubnet(ctx, provider, createdID)
	if err != nil {
		return nil, fmt.Errorf("Error waiting for instance (%s) to be created: %s", *res.JSON201.Id, err)
	}

	if mapPublicIPOnLaunch {
		if _, err = UpdateSubnetAttributes(ctx, provider, createdID, true); err != nil {
			return nil, fmt.Errorf("failed to update MapPublicIPOnLaunch: %w", err)
		}
	}

	if len(tags) > 0 {
		if err = CreateTags(ctx, provider, createdID, tags); err != nil {
			return nil, fmt.Errorf("failed to update tags: %w", err)
		}
	}

	resRead, err := ReadSubnet(ctx, provider, createdID)
	if err != nil {
		return nil, fmt.Errorf("failed to read subnet: %w", err)
	}

	return resRead, nil
}

func RetryReadSubnet(ctx context.Context, provider services.IProvider, id string) (*numspot.Subnet, error) {
	read, err := utils2.RetryReadUntilStateValid(
		ctx,
		id,
		provider.GetSpaceID(),
		[]string{pending},
		[]string{available},
		provider.GetNumspotClient().ReadSubnetsByIdWithResponse,
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

func ReadSubnet(ctx context.Context, provider services.IProvider, id string) (*numspot.Subnet, error) {
	res, err := provider.GetNumspotClient().ReadSubnetsByIdWithResponse(ctx, provider.GetSpaceID(), id)
	if err != nil {
		return nil, err
	}

	if res.JSON200 == nil {
		return nil, fmt.Errorf("got %s", res.Status())
	}

	return res.JSON200, nil
}

func UpdateSubnetAttributes(ctx context.Context, provider services.IProvider, id string, mapPublicIpOnLaunch bool) (*numspot.Subnet, error) {
	res, err := provider.GetNumspotClient().UpdateSubnetWithResponse(ctx, provider.GetSpaceID(), id, numspot.UpdateSubnetJSONRequestBody{
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

func UpdateSubnetTags(ctx context.Context, provider services.IProvider, id string, stateTags []numspot.ResourceTag, planTags []numspot.ResourceTag) (*numspot.Subnet, error) {
	if err := UpdateResourceTags(ctx, provider, stateTags, planTags, id); err != nil {
		return nil, err
	}
	return ReadSubnet(ctx, provider, id)
}

func DeleteSubnet(ctx context.Context, provider services.IProvider, id string) error {
	err := utils2.RetryDeleteUntilResourceAvailable(ctx, provider.GetSpaceID(), id, provider.GetNumspotClient().DeleteSubnetWithResponse)
	if err != nil {
		return err
	}

	return nil
}
