package fastly

import (
	"fmt"
	"log"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccFastlyServiceV1_wasm_package_basic(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.%s.com", name)

	wp1 := gofastly.Package{
		Metadata: gofastly.PackageMetadata{
			Name:        "package",
			Description: "eadsgsadg",
			Authors:     []string{"sgsfagasgfs"},
			Language:    "rust",
			Size:        0,
			HashSum:     "",
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1PackageConfig(name, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_wasm_v1.foo", &service),
					testAccCheckFastlyServiceV1PackageAttributes(&service, &wp1),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_wasm_v1.foo", "package.#", "1"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1PackageAttributes(service *gofastly.ServiceDetail, wasmPackage *gofastly.Package) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wp, err := conn.GetPackage(&gofastly.GetPackageInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		log.Printf("[DEBUG] Package = %#v\n", wp)

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Package for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if wp.Metadata.Size != wp.Metadata.Size {
			return fmt.Errorf("[ERR] Error looking up Package for (%s), version (%d): %s", service.Name, service.ActiveVersion.Number, err)
		}

		return nil
	}
}

func testAccServiceV1PackageConfig(name string, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_wasm_v1" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-loggly-logging"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  package {
    filename = "/Users/guy/workspace/terraform-provider-fastly/fastly/test_fixtures/package/test.tar.gz"
  }
  force_destroy = true
}
`, name, domain)
}
