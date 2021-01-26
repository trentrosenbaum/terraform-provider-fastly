package fastly

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func init() {
	resource.AddTestSweepers("fastly_tls_subscription", &resource.Sweeper{
		Name: "fastly_tls_subscription",
		F:    testSweepTLSSubscription,
	})
}

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

func testSweepTLSSubscription(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return err
	}

	subscriptions, err := client.ListTLSSubscriptions(&fastly.ListTLSSubscriptionsInput{PageSize: 1000})
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		for _, domain := range subscription.TLSDomains {
			if !strings.HasPrefix(domain.ID, testResourcePrefix) {
				continue
			}

			err = client.DeleteTLSSubscription(&fastly.DeleteTLSSubscriptionInput{
				ID: subscription.ID,
			})
			if err != nil {
				return err
			}
			break
		}
	}

	return nil
}
