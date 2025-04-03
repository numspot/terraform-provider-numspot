package core

//
//import (
//	"context"
//	"terraform-provider-numspot/internal/client"
//	"terraform-provider-numspot/internal/sdk/api"
//	"terraform-provider-numspot/internal/utils"
//)
//
//var (
//	directLinkInterfaceStates       = []string{creating, pending}
//	directLinkInterfaceTargetStates = []string{available}
//)
//
//func CreateDirectLinkInterface(ctx context.Context, provider *client.NumSpotSDK, createRequest api.CreateDirectLinkInterfaceRequest) (*api.DirectLinkInterface, error) {
//	spaceID := provider.SpaceID
//
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	var retryCreateResponse *api.CreateDirectLinkInterfaceResponse
//	if retryCreateResponse, err = utils.RetryCreateUntilResourceAvailableWithBody(ctx, spaceID, createRequest, numspotClient.CreateDirectLinkInterfaceWithResponse); err != nil {
//		return nil, err
//	}
//
//	return retryCreateResponse.JSON201, nil
//}
//
//func DeleteDirectLinkInterface(ctx context.Context, provider *client.NumSpotSDK) (*api.DirectLinkInterface, error) {
//	spaceID := provider.SpaceID
//
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//}
//
//func RetryReadDirectLinkInterface(ctx context.Context, provider *client.NumSpotSDK, directLinkInterfaceID string) {
//	numspotClient, err := provider.GetClient(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	read, err := utils.RetryReadUntilStateValid(ctx, directLinkInterfaceID, provider.SpaceID, imagePendingStates, imageTargetStates, numspotClient.ReadImagesByIdWithResponse)
//	if err != nil {
//		return nil, err
//	}
//
//}
//
//func ReadDirectLinkInterfaces() {
//
//}
//
//func ReadDirectLinkInterface() {
//
//}
