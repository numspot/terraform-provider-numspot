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
	"terraform-provider-numspot/internal/sdk/objectstorage"
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
	OsClient              *objectstorage.ClientWithResponses
	S3Creds               *S3Client
	HTTPClient            *http.Client
	SpaceID               api.SpaceId
	ClientID              uuid.UUID
	ClientSecret          string
	Host                  string
	HostOs                string
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

func WithHostOs(hostOs string) Option {
	return func(s *NumSpotSDK) error {
		s.HostOs = hostOs
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

	err := sdk.createClientAPI()
	if err != nil {
		return nil, err
	}

	err = sdk.createClientOs()
	if err != nil {
		return nil, err
	}

	err = sdk.authenticateUser(ctx)
	if err != nil {
		return nil, err
	}

	return sdk, nil
}

func isTokenExpired(expirationTime time.Time) bool {
	return time.Now().After(expirationTime)
}

func (s *NumSpotSDK) GetClient(ctx context.Context) (*api.ClientWithResponses, error) {
	if isTokenExpired(s.AccessTokenExpiration) {
		if err := s.authenticateUser(ctx); err != nil {
			return nil, fmt.Errorf("error while refreshing access token : %v", err)
		}
		s.AccessTokenExpiration = time.Now()
	}
	return s.Client, nil
}

func (s *NumSpotSDK) createClientAPI() error {
	requestEditor := api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add(UserAgentHeader, TerraformUserAgent)
		return nil
	})

	var err error
	s.Client, err = api.NewClientWithResponses(s.Host, s.newApiTransport(), requestEditor)
	if err != nil {
		return err
	}

	return nil
}

func (s *NumSpotSDK) createClientOs() error {
	newTransportOs := func(c *objectstorage.Client) error {
		if s.HTTPClient != nil {
			c.Client = s.HTTPClient
		} else {
			c.Client = &http.Client{
				Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			}
		}
		return nil
	}

	requestEditorOs := objectstorage.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add(UserAgentHeader, TerraformUserAgent)
		return nil
	})

	var err error
	s.OsClient, err = objectstorage.NewClientWithResponses(s.HostOs, newTransportOs, requestEditorOs)
	if err != nil {
		return err
	}

	return nil
}

func (s *NumSpotSDK) authenticateUser(ctx context.Context) error {
	basicAuth := buildBasicAuth(s.ClientID.String(), s.ClientSecret)

	response, err := s.Client.TokenWithFormdataBodyWithResponse(ctx, &api.TokenParams{Authorization: &basicAuth},
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

	s.Client, err = api.NewClientWithResponses(s.Host, s.newApiTransport(), api.WithRequestEditorFn(bearerProvider.Intercept))
	if err != nil {
		return err
	}

	err = s.setupS3Client(ctx, response)
	if err != nil {
		return err
	}

	return nil
}

func (s *NumSpotSDK) setupS3Client(ctx context.Context, response *api.TokenResponse) error {
	res, err := s.Client.ConvertTokenWithResponse(ctx, api.ConvertTokenJSONRequestBody{Token: response.JSON200.AccessToken})
	if err != nil || res.StatusCode() != 200 {
		return err
	}

	s.S3Creds = &S3Client{
		Ak:      res.JSON200.Ak,
		Sk:      res.JSON200.Sk,
		Service: ServiceS3,
		Region:  RegionS3,
	}

	return err
}

func buildBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (s *NumSpotSDK) newApiTransport() func(c *api.Client) error {
	return func(c *api.Client) error {
		if s.HTTPClient != nil {
			c.Client = s.HTTPClient
		} else {
			c.Client = &http.Client{
				Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
			}
		}
		return nil
	}
}
