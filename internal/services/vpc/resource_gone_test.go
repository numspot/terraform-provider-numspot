package vpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/client"
	"github.com/numspot/terraform-provider-numspot/internal/services/vpc/resource_vpc"
)

// vpcGoneServer returns a test server serving the Numspot IAM endpoints and a
// VPC read endpoint whose response is driven by the given content type, status
// and body. The SDK construction (token + convert-token) and the resource Read
// flow all exercise this stub.
func vpcGoneServer(t *testing.T, contentType string, vpcStatus int, vpcBody string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/iam/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"FAKE","expires_in":3600}`))
		case r.Method == http.MethodPut && r.URL.Path == "/iam/token/convert":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ak":"A","sk":"S"}`))
		default:
			w.Header().Set("Content-Type", contentType)
			w.WriteHeader(vpcStatus)
			_, _ = w.Write([]byte(vpcBody))
		}
	}))
	t.Cleanup(srv.Close)

	return srv
}

// newVPCResourceWithSDK returns a configured numspot_vpc resource whose
// provider points at the given test server, exactly like the provider wiring
// does at runtime (framework -> resource -> core -> api client).
func newVPCResourceWithSDK(t *testing.T, srvURL, spaceID string) resource.Resource {
	t.Helper()

	ctx := context.Background()
	sdk, err := client.NewNumspotSDK(ctx,
		client.WithHost(srvURL),
		client.WithHostOs(srvURL),
		client.WithSpaceID(spaceID),
		client.WithClientID(uuid.NewString()),
		client.WithClientSecret("secret"),
	)
	require.NoError(t, err)

	r := NewVPCResource()
	configureReq := resource.ConfigureRequest{ProviderData: sdk}
	configureResp := &resource.ConfigureResponse{}
	configurator, ok := r.(resource.ResourceWithConfigure)
	require.True(t, ok, "vpc resource must implement resource.ResourceWithConfigure")
	configurator.Configure(ctx, configureReq, configureResp)
	require.False(t, configureResp.Diagnostics.HasError())

	return r
}

// vpcStateRaw builds a raw state object matching the VPC resource schema,
// deriving every type from the schema itself so equality with the framework
// expectations is guaranteed.
func vpcStateRaw(t *testing.T, s schema.Schema, id string) tftypes.Value {
	t.Helper()

	ctx := context.Background()
	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok, "resource schema should render an object type")

	values := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		val := tftypes.NewValue(attrType, nil)
		switch name {
		case "id":
			val = tftypes.NewValue(attrType, id)
		case "ip_range":
			val = tftypes.NewValue(attrType, "10.0.0.0/16")
		case "state":
			val = tftypes.NewValue(attrType, "available")
		case "tenancy":
			val = tftypes.NewValue(attrType, "default")
		}
		values[name] = val
	}

	return tftypes.NewValue(objType, values)
}

// readVPCResource runs the generated Read method end to end (framework request
// -> resource Read -> core.ReadVPC -> api client -> HTTP stub).
func readVPCResource(t *testing.T, srvURL, spaceID string) *resource.ReadResponse {
	t.Helper()

	ctx := context.Background()
	r := newVPCResourceWithSDK(t, srvURL, spaceID)
	s := resource_vpc.VpcResourceSchema(ctx)
	raw := vpcStateRaw(t, s, "vpc-01234567")

	request := resource.ReadRequest{State: tfsdk.State{Schema: s, Raw: raw}}
	response := &resource.ReadResponse{State: tfsdk.State{Schema: s, Raw: raw}}
	r.Read(ctx, request, response)

	return response
}

func TestVPCResourceReadRemovesGoneResource(t *testing.T) {
	t.Parallel()

	// The Numspot API answers 403 (IAM masking) or 404 (not found) when a
	// resource cannot be seen: the Read must declare it absent
	// (State.RemoveResource) instead of failing the whole plan.
	tests := []struct {
		name        string
		contentType string
		status      int
		body        string
	}{
		{
			name:        "403 forbidden access",
			contentType: "application/json",
			status:      http.StatusForbidden,
			body:        `{"title":"Forbidden Access","type":"urn:numspot:errors#forbidden_access"}`,
		},
		{
			name:        "404 not found",
			contentType: "application/json",
			status:      http.StatusNotFound,
			body:        `{"title":"Not Found","type":"urn:numspot:errors#not_found"}`,
		},
		{
			name:        "404 with unparsable body",
			contentType: "text/plain",
			status:      http.StatusNotFound,
			body:        "not json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			srv := vpcGoneServer(t, tt.contentType, tt.status, tt.body)
			resp := readVPCResource(t, srv.URL, uuid.NewString())

			assert.False(t, resp.Diagnostics.HasError(),
				"a gone resource must not fail the Read: %v", resp.Diagnostics)
			assert.True(t, resp.State.Raw.IsNull(),
				"a gone resource must be removed from state")
		})
	}
}

func TestVPCResourceReadKeepsHardErrorOnOtherStatuses(t *testing.T) {
	t.Parallel()

	srv := vpcGoneServer(t, "application/json", http.StatusInternalServerError,
		`{"title":"Internal Server Error","type":"urn:numspot:errors#internal_error"}`)
	resp := readVPCResource(t, srv.URL, uuid.NewString())

	assert.True(t, resp.Diagnostics.HasError(), "a server error must stay a hard error")
	assert.False(t, resp.State.Raw.IsNull(), "the prior state must be kept on a hard error")
}

func TestVPCResourceReadKeepsStateOn200(t *testing.T) {
	t.Parallel()

	// Anti-false-positive control: with the API answering 200, the state must
	// be preserved and refreshed, proving the gone path is not a false positive.
	body, err := json.Marshal(map[string]any{
		"id":      "vpc-01234567",
		"ipRange": "10.0.0.0/16",
		"state":   "available",
		"tenancy": "default",
		"tags":    []map[string]string{{"key": "env", "value": "test"}},
	})
	require.NoError(t, err)

	srv := vpcGoneServer(t, "application/json", http.StatusOK, string(body))
	resp := readVPCResource(t, srv.URL, uuid.NewString())

	assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)
	assert.False(t, resp.State.Raw.IsNull())

	var state resource_vpc.VpcModel
	diags := resp.State.Get(context.Background(), &state)
	assert.False(t, diags.HasError())
	assert.Equal(t, "vpc-01234567", state.Id.ValueString())
	assert.Equal(t, "available", state.State.ValueString())
	assert.Equal(t, "10.0.0.0/16", state.IpRange.ValueString())
}
