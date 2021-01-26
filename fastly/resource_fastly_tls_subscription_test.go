package fastly

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccResourceFastlyTLSSubscription(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.com", name)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceFastlyTLSSubscriptionConfig(name, domain),
			},
		},
	})
}

func testAccResourceFastlyTLSSubscriptionConfig(name, domain string) string {
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
resource "fastly_tls_subscription" "subject" {
  domains = [for domain in fastly_service_v1.test.domain : domain.name]
  certificate_authority = "lets-encrypt"
}
`, name, domain)
}
