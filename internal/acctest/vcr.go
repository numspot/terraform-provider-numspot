package acctest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const fixturesPath = "../test"

// scopePlaceholderUUID replaces real space/organisation UUIDs recorded in
// cassette request paths so that cassettes can be replayed with any dummy
// UUID configured in NUMSPOT_SPACE_ID / NUMSPOT_ORGANISATION_ID.
const scopePlaceholderUUID = "00000000-0000-0000-0000-000000000000"

var (
	spacePathRegex        = regexp.MustCompile(`(/spaces/)([0-9a-fA-F-]{36})([/?]|$)`)
	organisationPathRegex = regexp.MustCompile(`(/organisations/)([0-9a-fA-F-]{36})([/?]|$)`)
)

// normalizeScopeIDsInString replaces space/organisation UUIDs in URL paths
// with the fixed placeholder.
func normalizeScopeIDsInString(s string) string {
	s = spacePathRegex.ReplaceAllString(s, "${1}"+scopePlaceholderUUID+"${3}")
	s = organisationPathRegex.ReplaceAllString(s, "${1}"+scopePlaceholderUUID+"${3}")
	return s
}

// extractScopeIDs returns the space/organisation UUIDs found in s.
func extractScopeIDs(s string) []string {
	var ids []string
	for _, re := range []*regexp.Regexp{spacePathRegex, organisationPathRegex} {
		for _, m := range re.FindAllStringSubmatch(s, -1) {
			ids = append(ids, m[2])
		}
	}
	return ids
}

// normalizeKnownScopeIDs replaces the given real scope UUIDs with the fixed
// placeholder wherever they appear. This is used on request/response bodies
// in case a space or organisation ID appears inside JSON.
func normalizeKnownScopeIDs(s string, ids []string) string {
	for _, id := range ids {
		s = strings.ReplaceAll(s, id, scopePlaceholderUUID)
	}
	return s
}

func getTestFilePath(name, pkgFolder, suffix string) string {
	// special chars to ignore
	specialChars := regexp.MustCompile(`[\\?%*:|"<>. ]`)

	fileName := strings.ReplaceAll(name, "_", "-")
	fileName = specialChars.ReplaceAllLiteralString(fileName, "") + suffix

	return filepath.Join(pkgFolder, "testdata", fileName)
}

func newRecorder(vcrMode, testName, pkgPath string) (*recorder.Recorder, *http.Client, error) {
	var recorderMode recorder.Mode

	switch vcrMode {
	case "record":
		recorderMode = recorder.ModeRecordOnly
	case vcrModeReplay:
		recorderMode = recorder.ModeReplayOnly
	default:
		recorderMode = recorder.ModePassthrough
	}
	cassettePath := getTestFilePath(testName, pkgPath, ".cassette")
	// Setup recorder and scw client
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               recorderMode,
		SkipRequestLatency: true,
		RealTransport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create recorder: %w", err)
	}

	r.SetMatcher(cassetteMatcher)

	// remove sensitive information from the cassette
	r.AddHook(cassetteSanitizer, recorder.BeforeSaveHook)

	client := r.GetDefaultClient()
	return r, client, nil
}

func cassetteSanitizer(i *cassette.Interaction) error {
	// In order to avoid unnecessary diff after each cassette update:
	// We set request duration to fixed  5ms for all requests in all cassettes
	// we don't save date response header on cassettes
	i.Response.Duration = 5 * time.Millisecond
	delete(i.Response.Headers, "Date")

	delete(i.Request.Headers, "Authorization")
	delete(i.Request.Form, "client_id")
	delete(i.Request.Form, "client_secret")

	// Normalize space/organisation IDs before extracting them, so that a fresh
	// cassette recorded with one account can be replayed with another.
	scopeIDs := extractScopeIDs(i.Request.URL)
	i.Request.URL = normalizeScopeIDsInString(i.Request.URL)

	// remove plain body containing client_id and client_secret for iam/token endpoint
	if strings.HasSuffix(i.Request.URL, "iam/token") {
		i.Request.Body = ""
		i.Response.Body = `{"access_token": "FAKE_TOKEN","expires_in": 3600}`
	}
	// sanitize token/convert endpoint - mask tokens in request and response
	if strings.HasSuffix(i.Request.URL, "iam/token/convert") {
		i.Request.Body = `{"token":"FAKE_TOKEN"}`
		i.Response.Body = `{"ak":"FAKE_ACCESS_KEY","sk":"FAKE_SECRET_KEY"}`
	}

	// Sanitize certificate and private key in request body
	i.Request.Body = sanitizeCertificateAndKey(i.Request.Body)
	// Sanitize keypair in request and response bodies
	i.Request.Body = sanitizeKeypairInBody(i.Request.Body)
	i.Response.Body = sanitizeKeypairInBody(i.Response.Body)

	// Normalize any space/organisation IDs that appear inside request/response bodies.
	i.Request.Body = normalizeKnownScopeIDs(i.Request.Body, scopeIDs)
	i.Response.Body = normalizeKnownScopeIDs(i.Response.Body, scopeIDs)

	return nil
}

// sanitizeCertificateAndKey replaces certificate and private key content with placeholders
func sanitizeCertificateAndKey(body string) string {
	// Replace certificate body (-----BEGIN CERTIFICATE----- ... -----END CERTIFICATE-----)
	certRegex := regexp.MustCompile(`-{5}BEGIN CERTIFICATE-{5}[\s\S]*?-{5}END CERTIFICATE-{5}`) // gitleaks:allow
	body = certRegex.ReplaceAllString(body, "SERVER_CERTIFICATE")

	// Replace private key (-----BEGIN PRIVATE KEY----- ... -----END PRIVATE KEY-----) // gitleaks:allow
	privateKeyRegex := regexp.MustCompile(`-{5}BEGIN PRIVATE KEY-{5}[\s\S]*?-{5}END PRIVATE KEY-{5}`) // gitleaks:allow
	body = privateKeyRegex.ReplaceAllString(body, "SERVER_PRIVATE_KEY")

	// Also handle RSA private keys
	rsaKeyRegex := regexp.MustCompile(`-----BEGIN RSA PRIVATE KEY-----[\s\S]*?-----END RSA PRIVATE KEY-----`) // gitleaks:allow
	body = rsaKeyRegex.ReplaceAllString(body, "SERVER_PRIVATE_KEY")

	return body
}

func sanitizeKeypairInBody(body string) string {
	rsaKeyRegex := regexp.MustCompile(`-----BEGIN RSA PRIVATE KEY-----[\s\S]*?-----END RSA PRIVATE KEY-----`) // gitleaks:allow
	body = rsaKeyRegex.ReplaceAllString(body, "KEYPAIR_PRIVATE_KEY")

	openSSHKeyRegex := regexp.MustCompile(`-----BEGIN OPENSSH PRIVATE KEY-----[\s\S]*?-----END OPENSSH PRIVATE KEY-----`) // gitleaks:allow
	body = openSSHKeyRegex.ReplaceAllString(body, "KEYPAIR_PRIVATE_KEY")

	ecKeyRegex := regexp.MustCompile(`-----BEGIN EC PRIVATE KEY-----[\s\S]*?-----END EC PRIVATE KEY-----`) // gitleaks:allow
	body = ecKeyRegex.ReplaceAllString(body, "KEYPAIR_PRIVATE_KEY")

	privateKeyRegex := regexp.MustCompile(`-{5}BEGIN PRIVATE KEY-{5}[\s\S]*?-{5}END PRIVATE KEY-{5}`) // gitleaks:allow
	body = privateKeyRegex.ReplaceAllString(body, "KEYPAIR_PRIVATE_KEY")

	publicKeyRegex := regexp.MustCompile(`(ssh-ed25519|ssh-rsa|ssh-dss|ecdsa-sha2-[a-z0-9-]+)\s+[A-Za-z0-9+/=]+(?: [^"\n]*)?`) // gitleaks:allow
	body = publicKeyRegex.ReplaceAllString(body, "KEYPAIR_PUBLIC_KEY")

	return body
}

// cassetteBodyMatcher is a custom matcher that will juste check equivalence of request bodies
func cassetteBodyMatcher(actualRequest *http.Request, cassetteRequest *cassette.Request) bool {
	if actualRequest.Body == nil || actualRequest.ContentLength == 0 {
		return cassetteRequest.Body == ""
	}

	actualBody, err := actualRequest.GetBody()
	if err != nil {
		tflog.Error(context.Background(), fmt.Errorf("cassette body matcher: failed to copy actualRequest body: %w", err).Error())
		return false
	}
	actualRawBody, err := io.ReadAll(actualBody)
	if err != nil {
		tflog.Error(context.Background(), fmt.Errorf("cassette body matcher: failed to read actualRequest body: %w", err).Error())
		return false
	}

	actualBodySanitized := sanitizeCertificateAndKey(string(actualRawBody))
	actualBodySanitized = sanitizeKeypairInBody(actualBodySanitized)

	// Normalize any space/organisation IDs that appear inside JSON bodies so
	// a cassette recorded with one account can be replayed with another.
	scopeIDs := extractScopeIDs(cassetteRequest.URL)
	actualBodySanitized = normalizeKnownScopeIDs(actualBodySanitized, scopeIDs)
	cassetteBody := normalizeKnownScopeIDs(cassetteRequest.Body, scopeIDs)

	if actualBodySanitized == cassetteBody {
		return true
	}

	actualJSON := make(map[string]interface{})
	cassetteJSON := make(map[string]interface{})

	err = json.Unmarshal([]byte(actualBodySanitized), &actualJSON)
	if err != nil {
		tflog.Error(context.Background(), fmt.Errorf("cassette body matcher: failed to parse json body: %w", err).Error())
		return false
	}

	err = json.Unmarshal([]byte(cassetteBody), &cassetteJSON)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(cassetteJSON, actualJSON)
}

// cassetteMatcher is a custom matcher that check equivalence of a played request against a recorded one
//
//nolint:gocritic // go-vcr MatcherFunc requires a value cassette.Request parameter
func cassetteMatcher(actual *http.Request, expected cassette.Request) bool {
	expectedURL, _ := url.Parse(expected.URL)
	actualURL := actual.URL

	// Normalize space/organisation IDs so cassettes recorded with one account
	// can be replayed with another.
	expectedPath := normalizeScopeIDsInString(expectedURL.Path)
	actualPath := normalizeScopeIDsInString(actualURL.Path)

	match := actual.Method == expected.Method &&
		actualPath == expectedPath &&
		actualURL.RawQuery == expectedURL.RawQuery
	if strings.Contains(actualPath, "iam/token") {
		return match
	}

	return match && cassetteBodyMatcher(actual, &expected)
}
