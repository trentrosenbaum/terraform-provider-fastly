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
	key, cert, cert2, err := generateKeyAndMultipleCerts(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)
	cert2 = strings.ReplaceAll(cert2, "\n", `\n`)

	name := fmt.Sprintf("tf-test-%s", acctest.RandString(20))
	updatedName := fmt.Sprintf("tf-test-%s", acctest.RandString(20))

	resourceName := "fastly_tls_activation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccFastlyTLSActivationCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSActivationBasicConfig(name, name, key, name, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_id"),
					resource.TestCheckResourceAttr(resourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					testAccFastlyTLSActivationCheckExists(resourceName),
				),
			},
			{
				Config: testAccFastlyTLSActivationBasicConfig(name, name, key, updatedName, cert2, domain),
				Check:  testAccFastlyTLSActivationCheckExists(resourceName),
			},
		},
	})
}

func testAccFastlyTLSActivationBasicConfig(serviceName, keyName, key, certName, cert, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "test" {
  name = "%s"

  domain {
    name = "%s"
  }

  backend {
    address = "127.0.0.1"
    name    = "localhost"
  }

  force_destroy = true
}

resource "fastly_tls_private_key" "test" {
  key_pem = "%s"
  name = "%s"
}

resource "fastly_tls_certificate" "test" {
  certificate_body = "%s"
  name = "%s"
  depends_on = [fastly_tls_private_key.test]
}

resource "fastly_tls_activation" "test" {
  certificate_id = fastly_tls_certificate.test.id
  domain = "%s"
  depends_on = [fastly_service_v1.test]
}
`, serviceName, domain, key, keyName, cert, certName, domain)
}

func testAccFastlyTLSActivationCheckExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		conn := testAccProvider.Meta().(*FastlyClient).conn

		r := state.RootModule().Resources[resourceName]
		_, err := conn.GetTLSActivation(&fastly.GetTLSActivationInput{
			ID: r.Primary.ID,
		})
		return err
	}
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
