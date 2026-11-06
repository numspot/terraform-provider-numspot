package vpc

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/services/vpc/resource_vpc"
)

const (
	// vpcCreate201Body is the 201 payload the API returns right after a VPC
	// creation (state still provisioning).
	vpcCreate201Body = `{"id":"vpc-01234567","ipRange":"10.0.0.0/16","state":"pending","tenancy":"default"}`
	// vpcRead200Body is the read payload once the VPC reached its target state.
	vpcRead200Body = `{"id":"vpc-01234567","ipRange":"10.0.0.0/16","state":"available","tenancy":"default","tags":[{"key":"env","value":"test"}]}`
	// vpc403Body is the RFC 9457 problem the API returns on a forbidden create.
	vpc403Body = `{"title":"Forbidden Access","type":"urn:numspot:errors#forbidden_access"}`
)

// vpcCreateServer returns a test server serving the Numspot IAM endpoints, a
// VPC create endpoint (POST) answering createStatus/createBody, and any other
// call (the post-create read) answering readBody with 200. The SDK
// construction (token + convert-token) and the resource Create flow all
// exercise this stub.
func vpcCreateServer(t *testing.T, createStatus int, createBody, readBody string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/iam/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"FAKE","expires_in":3600}`))
		case r.Method == http.MethodPut && r.URL.Path == "/iam/token/convert":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ak":"A","sk":"S"}`))
		case r.Method == http.MethodPost:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(createStatus)
			_, _ = w.Write([]byte(createBody))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(readBody))
		}
	}))
	t.Cleanup(srv.Close)

	return srv
}

// createVPCResource runs the generated Create method end to end (framework
// request -> resource Create -> core.CreateVPC -> api client -> HTTP stub).
func createVPCResource(t *testing.T, srvURL, spaceID string) *resource.CreateResponse {
	t.Helper()

	ctx := context.Background()
	r := newVPCResourceWithSDK(t, srvURL, spaceID)
	s := resource_vpc.VpcResourceSchema(ctx)

	objType, ok := s.Type().TerraformType(ctx).(tftypes.Object)
	require.True(t, ok, "resource schema should render an object type")

	values := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		val := tftypes.NewValue(attrType, nil)
		if name == "ip_range" {
			val = tftypes.NewValue(attrType, "10.0.0.0/16")
		}
		values[name] = val
	}

	request := resource.CreateRequest{
		Plan: tfsdk.Plan{Schema: s, Raw: tftypes.NewValue(objType, values)},
	}
	response := &resource.CreateResponse{State: tfsdk.State{Schema: s}}
	r.Create(ctx, request, response)

	return response
}

func TestVPCResourceCreateFailsCleanlyOnForbidden(t *testing.T) {
	t.Parallel()

	// Regression: the API answering 403 (quota reached, IAM forbidden) used to
	// stop the create retry loop as a success; the caller then panicked
	// dereferencing a nil JSON201 and crashed the whole plugin. The create
	// must surface the API error as a diagnostic instead.
	srv := vpcCreateServer(t, http.StatusForbidden, vpc403Body, "")
	resp := createVPCResource(t, srv.URL, uuid.NewString())

	assert.True(t, resp.Diagnostics.HasError(), "a forbidden create must fail the plan")

	for _, d := range resp.Diagnostics.Errors() {
		assert.Contains(t, d.Summary()+": "+d.Detail(), "Forbidden Access",
			"the API error message must reach the user")
	}
}

func TestVPCResourceCreateSucceedsOn201(t *testing.T) {
	// Anti-regression control: with the API answering 201 then a 200 read,
	// the create keeps working after the retry stop-list change.
	//
	// Not parallel: t.Setenv must not race with parallel tests.
	t.Setenv("RETRY_BACKOFF", "1ms")

	srv := vpcCreateServer(t, http.StatusCreated, vpcCreate201Body, vpcRead200Body)
	resp := createVPCResource(t, srv.URL, uuid.NewString())

	assert.False(t, resp.Diagnostics.HasError(), "%v", resp.Diagnostics)

	var state resource_vpc.VpcModel
	diags := resp.State.Get(context.Background(), &state)
	require.False(t, diags.HasError())
	assert.Equal(t, "vpc-01234567", state.Id.ValueString())
	assert.Equal(t, "available", state.State.ValueString())
	assert.Equal(t, "10.0.0.0/16", state.IpRange.ValueString())
}

func TestVPCResourceCreateRetriesOn429(t *testing.T) {
	// The API may rate-limit creates (429 Too Many Requests, quota
	// exhaustion): the create must be retried with backoff until it succeeds,
	// instead of failing the apply on the first throttled call.
	//
	// Not parallel: t.Setenv must not race with parallel tests.
	t.Setenv("RETRY_BACKOFF", "1ms")

	var createCalls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/iam/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"FAKE","expires_in":3600}`))
		case r.Method == http.MethodPut && r.URL.Path == "/iam/token/convert":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"ak":"A","sk":"S"}`))
		case r.Method == http.MethodPost:
			// First create call is throttled, the retry succeeds.
			if atomic.AddInt32(&createCalls, 1) == 1 {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"title":"Too Many Requests","type":"urn:numspot:errors#too_many_requests"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(vpcCreate201Body))
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(vpcRead200Body))
		}
	}))
	t.Cleanup(srv.Close)

	resp := createVPCResource(t, srv.URL, uuid.NewString())

	assert.False(t, resp.Diagnostics.HasError(),
		"a throttled create must be retried, not failed: %v", resp.Diagnostics)
	assert.GreaterOrEqual(t, atomic.LoadInt32(&createCalls), int32(2),
		"the create must have been retried after the 429")

	var state resource_vpc.VpcModel
	diags := resp.State.Get(context.Background(), &state)
	require.False(t, diags.HasError())
	assert.Equal(t, "vpc-01234567", state.Id.ValueString())
}
