package client

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/google/uuid"

	"github.com/numspot/terraform-provider-numspot/internal/sdk/api"
	"github.com/numspot/terraform-provider-numspot/internal/sdk/objectstorage"
	"github.com/numspot/terraform-provider-numspot/internal/utils"
)

var (
	// errInvalidAccessTokenExpiration is returned when the token endpoint returns
	// an expiration time that is too short.
	errInvalidAccessTokenExpiration = errors.New("error while retrieving access token expiration time. Invalid expiration time found")
	// errFailedToConvertTokenForS3Signing is returned when the token conversion
	// for S3 signing fails.
	errFailedToConvertTokenForS3Signing = errors.New("failed to convert token for S3 signing")
)

const (
	// UserAgentHeader handles user agent header.
	UserAgentHeader = "User-Agent"
	// TerraformUserAgent is the user agent sent with provider requests.
	TerraformUserAgent = "TERRAFORM-NUMSPOT"
	// Credentials handles credentials.
	Credentials = "client_credentials"
	serviceS3   = "s3"
	regionS3    = "eu-west-2"
)

// NumspotSDK is the Numspot API client.
type NumspotSDK struct {
	ID                    string
	Client                *api.ClientWithResponses
	OsClient              *objectstorage.ClientWithResponses
	SignFunc              objectstorage.RequestEditorFn
	HTTPClient            *http.Client
	SpaceID               api.SpaceId
	OrganisationID        api.OrganisationId
	ClientID              uuid.UUID
	ClientSecret          string
	Host                  string
	HostOs                string
	AccessTokenExpiration time.Time
}

// Option configures a NumspotSDK option.
type Option func(s *NumspotSDK) error

// WithHost sets the host option.
func WithHost(host string) Option {
	return func(s *NumspotSDK) error {
		s.Host = host
		return nil
	}
}

// WithSpaceID sets the space ID option.
func WithSpaceID(spaceID string) Option {
	return func(s *NumspotSDK) error {
		numspotSpaceID, err := uuid.Parse(spaceID)
		if err != nil {
			return utils.WrapErrorf(err, "failed to parse space ID %q", spaceID)
		}
		s.SpaceID = numspotSpaceID
		return nil
	}
}

// WithOrganisationID sets the organisation ID option.
func WithOrganisationID(organisationID string) Option {
	return func(s *NumspotSDK) error {
		if organisationID == "" {
			return nil
		}
		numspotOrganisationID, err := uuid.Parse(organisationID)
		if err != nil {
			return utils.WrapErrorf(err, "failed to parse organisation ID %q", organisationID)
		}
		s.OrganisationID = numspotOrganisationID
		return nil
	}
}

// WithClientID sets the client ID option.
func WithClientID(clientID string) Option {
	return func(s *NumspotSDK) error {
		clientUUID, err := uuid.Parse(clientID)
		if err != nil {
			return utils.WrapErrorf(err, "failed to parse client ID %q", clientID)
		}
		s.ClientID = clientUUID
		return nil
	}
}

// WithClientSecret sets the client secret option.
func WithClientSecret(clientSecret string) Option {
	return func(s *NumspotSDK) error {
		s.ClientSecret = clientSecret
		return nil
	}
}

// WithHTTPClient sets the HTTP client option.
func WithHTTPClient(client *http.Client) Option {
	return func(s *NumspotSDK) error {
		s.HTTPClient = client
		return nil
	}
}

// WithHostOs sets the host OS option.
func WithHostOs(hostOs string) Option {
	return func(s *NumspotSDK) error {
		s.HostOs = hostOs
		return nil
	}
}

// NewNumspotSDK returns a new NumspotSDK.
func NewNumspotSDK(ctx context.Context, options ...Option) (*NumspotSDK, error) {
	sdk := &NumspotSDK{
		ID:                    uuid.NewString(),
		AccessTokenExpiration: time.Now(),
	}
	for _, o := range options {
		if err := o(sdk); err != nil {
			return nil, utils.WrapError(err, "failed to apply SDK option")
		}
	}

	err := sdk.createClientAPI()
	if err != nil {
		return nil, utils.WrapError(err, "failed to create API client")
	}

	err = sdk.createClientOs()
	if err != nil {
		return nil, utils.WrapError(err, "failed to create object storage client")
	}

	err = sdk.authenticateUser(ctx)
	if err != nil {
		return nil, utils.WrapError(err, "failed to authenticate user")
	}

	return sdk, nil
}

func isTokenExpired(expirationTime time.Time) bool {
	return time.Now().After(expirationTime)
}

// GetClient returns the API client.
func (s *NumspotSDK) GetClient(ctx context.Context) (*api.ClientWithResponses, error) {
	if isTokenExpired(s.AccessTokenExpiration) {
		if err := s.authenticateUser(ctx); err != nil {
			return nil, fmt.Errorf("error while refreshing access token: %w", err)
		}
	}
	return s.Client, nil
}

func (s *NumspotSDK) createClientAPI() error {
	requestEditor := api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		req.Header.Add(UserAgentHeader, TerraformUserAgent)
		return nil
	})

	var err error
	s.Client, err = api.NewClientWithResponses(s.Host, s.newApiTransport(), requestEditor)
	if err != nil {
		return utils.WrapError(err, "failed to create API client")
	}

	return nil
}

func (s *NumspotSDK) createClientOs() error {
	newTransportOs := func(c *objectstorage.Client) error {
		if s.HTTPClient != nil {
			c.Client = s.HTTPClient
		} else {
			c.Client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					Proxy:           http.ProxyFromEnvironment,
				},
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
		return utils.WrapError(err, "failed to create object storage client")
	}

	return nil
}

func (s *NumspotSDK) authenticateUser(ctx context.Context) error {
	basicAuth := buildBasicAuth(s.ClientID.String(), s.ClientSecret)

	response, err := s.Client.TokenWithFormdataBodyWithResponse(ctx, &api.TokenParams{Authorization: &basicAuth},
		api.TokenReq{
			GrantType:    Credentials,
			ClientId:     &s.ClientID,
			ClientSecret: &s.ClientSecret,
		},
	)
	if err != nil {
		return utils.WrapError(err, "failed to retrieve access token")
	}
	err = utils.ParseHTTPError(response.Body, response.StatusCode())
	if err != nil {
		return fmt.Errorf("error while retrieving access token for client: %w", err)
	}

	expirationTimeMargin := 5 * 60 // Add a margin of 5 minutes to refresh the token
	var expirationTime int
	if response.JSON200.ExpiresIn > expirationTimeMargin {
		expirationTime = response.JSON200.ExpiresIn - expirationTimeMargin
	} else {
		return fmt.Errorf(
			"%w. Found %v but expected an expiration time higher than %v",
			errInvalidAccessTokenExpiration,
			response.JSON200.ExpiresIn,
			expirationTimeMargin,
		)
	}

	s.AccessTokenExpiration = time.Now().Add(time.Duration(expirationTime) * time.Second)

	bearerProvider, err := securityprovider.NewSecurityProviderBearerToken(response.JSON200.AccessToken)
	if err != nil {
		return utils.WrapError(err, "failed to create bearer token provider")
	}

	s.Client, err = api.NewClientWithResponses(s.Host, s.newApiTransport(), api.WithRequestEditorFn(bearerProvider.Intercept))
	if err != nil {
		return utils.WrapError(err, "failed to create authenticated API client")
	}

	err = s.setupS3Client(ctx, response)
	if err != nil {
		return utils.WrapError(err, "failed to set up S3 client")
	}

	return nil
}

func (s *NumspotSDK) setupS3Client(ctx context.Context, response *api.TokenResponse) error {
	res, err := s.Client.ConvertTokenWithResponse(ctx, api.ConvertTokenJSONRequestBody{Token: response.JSON200.AccessToken})
	if err != nil {
		return utils.WrapError(err, "failed to convert token for S3 signing")
	}
	if res.StatusCode() != 200 {
		return fmt.Errorf("%w: unexpected status code %d", errFailedToConvertTokenForS3Signing, res.StatusCode())
	}

	s.SignFunc = utils.SignRequest(serviceS3, regionS3, res.JSON200.Ak, res.JSON200.Sk)

	return nil
}

func buildBasicAuth(username, password string) string {
	auth := username + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

func (s *NumspotSDK) newApiTransport() func(c *api.Client) error {
	return func(c *api.Client) error {
		if s.HTTPClient != nil {
			c.Client = s.HTTPClient
		} else {
			c.Client = &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					Proxy:           http.ProxyFromEnvironment,
				},
			}
		}
		return nil
	}
}
