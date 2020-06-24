package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyServiceWASMV1_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	comment := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	versionComment := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceWASMV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceWASMV1Config(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceWASMV1Exists("fastly_service_wasm_v1.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "comment", "Managed by Terraform"),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "version_comment", ""),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "backend.#", "1"),
				),
			},

			{
				Config: testAccServiceWASMV1Config_basicUpdate(name, comment, versionComment, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceWASMV1Exists("fastly_service_wasm_v1.foo", &service),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "comment", comment),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "version_comment", versionComment),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "domain.#", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "backend.#", "1"),
				),
			},
		},
	})
}

func testAccCheckServiceWASMV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_wasm_v1" {
			continue
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing servcies when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckServiceWASMV1Exists(n string, service *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service ID is set")
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		latest, err := conn.GetServiceDetails(&gofastly.GetServiceInput{
			ID: rs.Primary.ID,
		})

		if err != nil {
			return err
		}

		*service = *latest

		return nil
	}
}

func testAccServiceWASMV1Config(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_wasm_v1" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  force_destroy = true
  activate = false
}`, name, domain)
}

func testAccServiceWASMV1Config_basicUpdate(name, comment, versionComment, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_wasm_v1" "foo" {
  name    = "%s"
  comment = "%s"
  version_comment = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  force_destroy = true
}`, name, comment, versionComment, domain)
}
