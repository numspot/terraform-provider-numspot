package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/provider"
)

// ProtoV6ProviderFactoriesType handles proto v6 provider factories type.
type ProtoV6ProviderFactoriesType = map[string]func() (tfprotov6.ProviderServer, error)

// AccTest handles acc test.
type AccTest struct {
	TestProvider ProtoV6ProviderFactoriesType
	Cleanup      func() error
}

const vcrModeReplay = "replay"

// NewAccTest returns an AccTest with a VCR-backed provider.
//
// VCR is always enabled. The mode is controlled by the VCR_MODE environment
// variable:
//   - "record": use the real API and record/overwrite the cassette;
//   - "replay" or unset: replay the recorded cassette (default);
//   - any other value: passthrough (real API, no cassette saved).
//
// When VCR_MODE is unset the default is "replay". If the cassette does not
// exist in replay mode, the test fails fast with an instructive message.
func vcrMode() string {
	mode := os.Getenv("VCR_MODE")
	if mode == "" {
		return vcrModeReplay
	}
	return mode
}

// NewAccTest returns a new acc test.
func NewAccTest(t *testing.T) AccTest {
	t.Helper()
	runParallel := os.Getenv("PARALLEL_TEST")
	if runParallel == "true" {
		t.Parallel()
	}

	mode := vcrMode()

	if mode == vcrModeReplay {
		cassettePath := getTestFilePath(t.Name(), fixturesPath, ".cassette") + ".yaml"
		if _, err := os.Stat(cassettePath); os.IsNotExist(err) {
			t.Fatalf("VCR replay cassette not found: %s\nSet VCR_MODE=record to record it, or ensure the cassette is committed.", cassettePath)
		}
	}

	rec, client, err := newRecorder(mode, t.Name(), fixturesPath)
	require.NoError(t, err)
	pr := provider.ProvideNumspotProviderWithHTTPClient(client)()

	return AccTest{
		TestProvider: ProtoV6ProviderFactoriesType{
			"numspot": providerserver.NewProtocol6WithError(pr),
		},
		Cleanup: rec.Stop,
	}
}
