package core

import (
	"context"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

// CreateServerCertificate creates server certificate.
func CreateServerCertificate(
	ctx context.Context,
	provider *client.NumspotSDK,
	numspotServerCertificateCreate *api.CreateServerCertificateJSONRequestBody,
) (*api.ServerCertificate, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	resp, err := numspotClient.CreateServerCertificateWithResponse(ctx, spaceID, *numspotServerCertificateCreate)
	if err != nil {
		return nil, utils.WrapError(err, "failed to create server certificate")
	}

	err = utils.ParseHTTPError(resp.Body, resp.StatusCode())
	if err != nil {
		return nil, utils.WrapError(err, "failed to parse HTTP error response")
	}

	return resp.JSON201, nil
}

// DeleteServerCertificate deletes server certificate.
func DeleteServerCertificate(ctx context.Context, provider *client.NumspotSDK, serverCertificateID string) error {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return utils.WrapError(err, "failed to get API client")
	}

	res, err := numspotClient.DeleteServerCertificate(ctx, spaceID, serverCertificateID)
	if err != nil {
		return utils.WrapError(err, "failed API client call")
	}
	defer func() { _ = res.Body.Close() }()

	return nil
}

// ReadServerCertificates reads server certificates.
func ReadServerCertificates(
	ctx context.Context,
	provider *client.NumspotSDK,
	serverCertificates *api.ReadServerCertificatesParams,
) (*[]api.ServerCertificate, error) {
	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	read, err := numspotClient.ReadServerCertificatesWithResponse(ctx, provider.SpaceID, serverCertificates)
	if err != nil {
		return nil, utils.WrapError(err, "failed to read server certificates")
	}

	if err = utils.ParseHTTPError(read.Body, read.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to read server certificates")
	}

	return read.JSON200.Items, nil
}

// ReadServerCertificate reads server certificate.
func ReadServerCertificate(ctx context.Context, provider *client.NumspotSDK, serverCertificateId string) (*api.ServerCertificate, error) {
	resp, err := ReadServerCertificates(ctx, provider, &api.ReadServerCertificatesParams{})
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

// UpdateServerCertificate updates server certificate.
func UpdateServerCertificate(
	ctx context.Context,
	provider *client.NumspotSDK,
	id string,
	body api.UpdateServerCertificateJSONRequestBody,
) (*api.ServerCertificate, error) {
	spaceID := provider.SpaceID

	numspotClient, err := provider.GetClient(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to get API client")
	}

	resp, err := numspotClient.UpdateServerCertificateWithResponse(ctx, spaceID, id, body)
	if err != nil {
		return nil, utils.WrapError(err, "failed to update server certificate")
	}

	if err = utils.ParseHTTPError(resp.Body, resp.StatusCode()); err != nil {
		return nil, utils.WrapError(err, "failed to update server certificate")
	}

	return ReadServerCertificate(ctx, provider, id)
}
