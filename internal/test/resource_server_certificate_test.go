//go:build acc

package test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/numspot/terraform-provider-numspot/internal/acctest"
)

func TestAccServerCertificateResource(t *testing.T) {
	certBody := acctest.GetEnvOrReplay(t, "SERVER_CERTIFICATE_ACCTEST", "SERVER_CERTIFICATE")
	privateKey := acctest.GetEnvOrReplay(t, "SERVER_PRIVATE_KEY_ACCTEST", "SERVER_PRIVATE_KEY")

	acct := acctest.NewAccTest(t)
	defer func() {
		err := acct.Cleanup()
		require.NoError(t, err)
	}()
	pr := acct.TestProvider

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: pr,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
resource "numspot_server_certificate" "server-certificate-acctest" {
  name        = %[1]q
  body        = %[2]q
  private_key = %[3]q
}
	`, "server-certificate-acctest", certBody, privateKey),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("numspot_server_certificate.server-certificate-acctest", "name", "server-certificate-acctest"),
				),
			},
		},
	})
}
