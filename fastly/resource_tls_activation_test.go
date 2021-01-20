package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAccFastlyTLSActivationBasic(t *testing.T) {
	domain := fmt.Sprintf("tf-test-%s.com", acctest.RandString(10))
	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFastlyTLSActivationCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSActivationBasicConfig(key, cert, domain),
			},
		},
	})
}

func testAccFastlyTLSActivationBasicConfig(key, cert, domain string) string {
	certificateName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_tls_private_key" "test" {
  key_pem = "%s"
  name = "%s"
}

resource "fastly_tls_certificate" "test" {
  certificate_body = "%s"
  name = "%s"
}

resource "fastly_tls_activation" "test" {
  certificate_id = fastly_tls_certificate.test.id
  domain = "%s"
}
`, key, certificateName, cert, certificateName, domain)
}

func testAccFastlyTLSActivationCheckDestroy(state *terraform.State) error {
	for _, resourceState := range state.RootModule().Resources {
		if resourceState.Type != "fastly_tls_activation" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		activations, err := conn.ListTLSActivations(&fastly.ListTLSActivationsInput{})
		if err != nil {
			return fmt.Errorf(
				"[WARN] Error listing TLS activations when deleting activation %s: %w",
				resourceState.Primary.ID,
				err,
			)
		}

		for _, activation := range activations {
			if activation.ID == resourceState.Primary.ID {
				return fmt.Errorf(
					"[WARN] Tried disabling TLS activation (%s) but was still found",
					resourceState.Primary.ID,
				)
			}
		}
	}
	return nil
}
