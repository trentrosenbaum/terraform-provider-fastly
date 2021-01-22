package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestAccDataSourceFastlyTLSActivationIdsBasic(t *testing.T) {
	domain := fmt.Sprintf("tf-test-%s.com", acctest.RandomWithPrefix("tf-test-"))
	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)
	key = strings.ReplaceAll(key, "\n", `\n`)
	cert = strings.ReplaceAll(cert, "\n", `\n`)

	datasourceName := "data.fastly_tls_activation_ids.test"
	activationDatasourceName := "data.fastly_tls_activation.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceFastlyTLSActivationIdsConfig(key, cert, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "ids.#", "1"),
					resource.TestCheckResourceAttrSet(activationDatasourceName, "certificate_id"),
					resource.TestCheckResourceAttrSet(activationDatasourceName, "configuration_id"),
					resource.TestCheckResourceAttr(activationDatasourceName, "domain", domain),
					resource.TestCheckResourceAttrSet(activationDatasourceName, "created_at"),
				),
			},
		},
	})
}

func testAccDataSourceFastlyTLSActivationIdsConfig(key, cert, domain string) string {
	name := acctest.RandomWithPrefix(testResourcePrefix)

	return fmt.Sprintf(
		`
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

data "fastly_tls_activation_ids" "test" {
  certificate_id = fastly_tls_activation.test.certificate_id
}

data "fastly_tls_activation" "test" {
  id = tolist(data.fastly_tls_activation_ids.test.ids)[0]
}
`,
		name,
		domain,
		key,
		name,
		cert,
		name,
		domain,
	)
}
