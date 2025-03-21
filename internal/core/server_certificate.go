package core

import (
	"context"
	"errors"
	"net/http"

	"terraform-provider-numspot/internal/client"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

func CreateServerCertificate(ctx context.Context, provider *client.NumSpotSDK, numSpotServerCertificateCreate api.CreateServerCertificateJSONRequestBody) (*api.ServerCertificate, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := numspotClient.CreateServerCertificateWithResponse(ctx, spaceID, numSpotServerCertificateCreate)
	if err != nil {
		return nil, err
	}

	err = utils.ParseHTTPError(resp.Body, resp.StatusCode())
	if err != nil {
		return nil, err
	}

	return resp.JSON201, nil
}

func DeleteServerCertificate(ctx context.Context, provider *client.NumSpotSDK, serverCertificateID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return err
	}

	_, err = numspotClient.DeleteServerCertificate(ctx, spaceID, serverCertificateID)
	if err != nil {
		return err
	}

	return nil
}

func ReadServerCertificates(ctx context.Context, provider *client.NumSpotSDK, serverCertificates *api.ReadServerCertificatesParams) (*[]api.ServerCertificate, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	read, err := numspotClient.ReadServerCertificatesWithResponse(ctx, provider.SpaceID, serverCertificates)
	if err != nil {
		return nil, err
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, err
	}

	return read.JSON200.Items, nil
}

func ReadServerCertificate(ctx context.Context, provider *client.NumSpotSDK, serverCertificateId string) (*api.ServerCertificate, error) {
	resp, err := ReadServerCertificates(ctx, provider, nil)
	if err != nil {
		return nil, err
	}

	var ret api.ServerCertificate
	if resp != nil {
		ll := len(*resp)
		stop := false
		for i := 0; ll > i && !stop; i++ {
			if *(*resp)[i].Name == serverCertificateId {
				ret = (*resp)[i]
				stop = true
			}
		}
	}

	return &ret, nil
}

func UpdateServerCertificate(ctx context.Context, provider *client.NumSpotSDK, id string, body api.UpdateServerCertificateJSONRequestBody) (*api.ServerCertificate, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := numspotClient.UpdateServerCertificateWithResponse(ctx, spaceID, id, body)
	if err != nil {
		return nil, err
	}

	if resp.Status() != string(rune(http.StatusNoContent)) {
		return nil, errors.New("implement error http parsing")
	}

	return ReadServerCertificate(ctx, provider, id)
}
