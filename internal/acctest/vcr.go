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

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const fixturesPath = "../test"

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
	case "replay":
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
		return nil, nil, err
	}

	r.SetMatcher(cassetteMatcher)

	// remove sensitive information from the cassette
	r.AddHook(cassetteSanitizer, recorder.BeforeSaveHook)

	return r, r.GetDefaultClient(), nil
}

func cassetteSanitizer(i *cassette.Interaction) error {
	delete(i.Request.Headers, "Authorization")
	delete(i.Request.Form, "client_id")
	delete(i.Request.Form, "client_secret")
	// remove plain body containing client_id and client_secret for iam/token endpoint
	if strings.HasSuffix(i.Request.URL, "iam/token") {
		i.Request.Body = ""
		i.Response.Body = `{"access_token": "just a fake token"}`
	}
	return nil
}

// cassetteBodyMatcher is a custom matcher that will juste check equivalence of request bodies
func cassetteBodyMatcher(actualRequest *http.Request, cassetteRequest cassette.Request) bool {
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

	if string(actualRawBody) == cassetteRequest.Body {
		return true
	}

	actualJSON := make(map[string]interface{})
	cassetteJSON := make(map[string]interface{})

	err = json.Unmarshal(actualRawBody, &actualJSON)
	if err != nil {
		tflog.Error(context.Background(), fmt.Errorf("cassette body matcher: failed to parse json body: %w", err).Error())
		return false
	}

	err = json.Unmarshal([]byte(cassetteRequest.Body), &cassetteJSON)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(cassetteJSON, actualJSON)
}

// cassetteMatcher is a custom matcher that check equivalence of a played request against a recorded one
func cassetteMatcher(actual *http.Request, expected cassette.Request) bool {
	expectedURL, _ := url.Parse(expected.URL)
	actualURL := actual.URL

	match := actual.Method == expected.Method &&
		actual.URL.Path == expectedURL.Path &&
		actualURL.RawQuery == expectedURL.RawQuery
	if strings.Contains(actualURL.Path, "iam/token") {
		return match
	}

	return match && cassetteBodyMatcher(actual, expected)
}
