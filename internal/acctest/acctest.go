package acctest

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/stretchr/testify/require"

	"gitlab.numspot.cloud/cloud/terraform-provider-numspot/internal/provider"
)

type ProtoV6ProviderFactoriesType = map[string]func() (tfprotov6.ProviderServer, error)

type AccTest struct {
	TestProvider ProtoV6ProviderFactoriesType
	Cleanup      func() error
}

// NewAccTest will return AccTest object encapsulating test provider instance and a cleanup function
//
// If vcrEnabled set to true it will return a provider instance with go-vcr http client and the cleanup
// will make reference to recorder.stop() func.
//
//	otherwise will return default provider and empty cleanup func.
//
// vcrMode possible values:
//
// "record": Use real API calls and record to cassettes files,
//
//	If cassette files exists wit will be overwritten.
//
// "replay": Use Mock API calls from the values recorded in the cassettes
//
//	and record new episodes if not present in the cassettes file
//
// # Any other value will use Pass through mode and use real API calls without saving to cassettes
//
// The same applies to VCR_MODE env variable, the only difference is that this mode will be used across all tests.
func NewAccTest(t *testing.T, vcrEnabled bool, vcrMode string) AccTest {
	t.Helper()

	// Determine mode based on the environment variable or passed parameters
	mode := os.Getenv("VCR_MODE")

	if mode == "" {
		// Return immediately when VCR is not enabled
		if !vcrEnabled {
			return AccTest{
				TestProvider: ProtoV6ProviderFactoriesType{
					"numspot": providerserver.NewProtocol6WithError(provider.New("0.1", false, nil)()),
				},
				Cleanup: func() error {
					return nil
				},
			}
		}

		// Set mode to vcrMode if no global VCR mode is set and VCR is enabled
		mode = vcrMode
	}

	// cwd is internal/provider, the path to acctest pkg is: ../acctest

	rec, client, err := newRecorder(mode, t.Name(), fixturesPath)
	require.NoError(t, err)
	pr := provider.New("0.1", false, client)()

	return AccTest{
		TestProvider: ProtoV6ProviderFactoriesType{
			"numspot": providerserver.NewProtocol6WithError(pr),
		},
		Cleanup: rec.Stop,
	}
}
