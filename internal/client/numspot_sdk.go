package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/google/uuid"
	"terraform-provider-numspot/internal/sdk/api"
	"terraform-provider-numspot/internal/utils"
)

const (
	UserAgentHeader    = "User-Agent"
	TerraformUserAgent = "TERRAFORM-NUMSPOT"
	Credentials        = "client_credentials"
	ServiceS3          = "s3"
	RegionS3           = "eu-west-2"
)

type S3Client struct {
	Service string
	Region  string
	Ak      string
	Sk      string
}

type NumSpotSDK struct {
	ID                    string
	Client                *api.ClientWithResponses
	HTTPClient            *http.Client
	SpaceID               api.SpaceId
	ClientID              uuid.UUID
	ClientSecret          string
	Host                  string
	AccessTokenExpiration time.Time
}

type Option func(s *NumSpotSDK) error

func WithHost(host string) Option {
	return func(s *NumSpotSDK) error {
		s.Host = host
		return nil
	}
}

func WithSpaceID(spaceID string) Option {
	return func(s *NumSpotSDK) error {
		numSpotSpaceID, err := uuid.Parse(spaceID)
		if err != nil {
			return err
		}
		s.SpaceID = numSpotSpaceID
		return nil
	}
}

func WithClientID(clientID string) Option {
	return func(s *NumSpotSDK) error {
		clientUUID, err := uuid.Parse(clientID)
		if err != nil {
			return err
		}
		s.ClientID = clientUUID
		return nil
	}
}

func WithClientSecret(clientSecret string) Option {
	return func(s *NumSpotSDK) error {
		s.ClientSecret = clientSecret
		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(s *NumSpotSDK) error {
		s.HTTPClient = client
		return nil
	}
}

func NewNumSpotSDK(ctx context.Context, options ...Option) (*NumSpotSDK, error) {
	sdk := &NumSpotSDK{
		ID:                    uuid.NewString(),
		AccessTokenExpiration: time.Now(),
	}
	for _, o := range options {
		if err := o(sdk); err != nil {
			return nil, err
		}
	}

	if err := sdk.AuthenticateUser(ctx); err != nil {
		return nil, err
	}

	return sdk, nil
}

func isTokenExpired(expirationTime time.Time) bool {
	return time.Now().After(expirationTime)
}

func (s *NumSpotSDK) GetClient(ctx context.Context) (*api.ClientWithResponses, error) {
	if isTokenExpired(s.AccessTokenExpiration) {
		if err := s.AuthenticateUser(ctx); err != nil {
			return nil, fmt.Errorf("error while refreshing access token : %v", err)
		}
		s.AccessTokenExpiration = time.Now()
	}
	return s.Client, nil
}

func (s *NumSpotSDK) AuthenticateUser(ctx context.Context) error {
	var err error
	// TODO can we activate secure transport ?
	newTransport := func(c *api.Client) error {
		if s.HTTPClient != nil {
			c.Client = s.HTTPClient
		} else {
			c.Client = &http.Client{
				Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			}
		}
		return nil
	}

	requestEditor := api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add(UserAgentHeader, TerraformUserAgent)
		return nil
	})

	numSpotClient, err := api.NewClientWithResponses(s.Host, newTransport, requestEditor)
	if err != nil {
		return err
	}

	basicAuth := buildBasicAuth(s.ClientID.String(), s.ClientSecret)

	response, err := numSpotClient.TokenWithFormdataBodyWithResponse(ctx, &api.TokenParams{Authorization: &basicAuth},
		api.TokenReq{
			GrantType:    Credentials,
			ClientId:     &s.ClientID,
			ClientSecret: &s.ClientSecret,
		},
	)
	if err != nil {
		return err
	}
	err = utils.ParseHTTPError(response.Body, response.StatusCode())
	if err != nil {
		return fmt.Errorf("error while retrieving access token for client : %v", err.Error())
	}

	expirationTimeMargin := 5 * 60 // Add a margin of 5 minutes to refresh the token
	var expirationTime int
	if response.JSON200.ExpiresIn > expirationTimeMargin {
		expirationTime = response.JSON200.ExpiresIn - expirationTimeMargin
	} else {
		return fmt.Errorf("error while retrieving access token expiration time. Invalid expiration time found. Found %v but expected and expiration time higher than %v", response.JSON200.ExpiresIn, expirationTimeMargin)
	}

	s.AccessTokenExpiration = time.Now().Add(time.Duration(expirationTime) * time.Second)

	bearerProvider, err := securityprovider.NewSecurityProviderBearerToken(response.JSON200.AccessToken)
	if err != nil {
		return err
	}

	s.Client, err = api.NewClientWithResponses(s.Host, newTransport, api.WithRequestEditorFn(bearerProvider.Intercept))
	if err != nil {
		return err
	}

	return nil
}

func buildBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}
