package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"strings"
	"testing"
)

func TestAccFastlyTLSPrivateKeyV1Create(t *testing.T) {
	key, _, err := generateKeyAndCert()
	if err != nil {
		t.Fatalf("Failed to generate private key: %v", err)
	}
	// TODO: string replacement in the resourceCreate function?
	key = strings.ReplaceAll(key, "\n", `\n`)

	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckPrivateKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyTLSPrivateKeyV1Config_simple_private_key(key, name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPrivateKeyExists("fastly_tls_private_key.foo"),
					resource.TestCheckResourceAttr("fastly_tls_private_key.foo", "name", name),
				),
			},
		},
	})
}

func testAccCheckPrivateKeyDestroy(_ *terraform.State) error {
	// TODO: implement check destroy
	return nil
}

func testAccFastlyTLSPrivateKeyV1Config_simple_private_key(key, name string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "foo" {
  key_pem = "%s"
  name    = "%s"
}`, key, name)
}

func testAccCheckPrivateKeyExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		res, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if res.Primary.ID == "" {
			return fmt.Errorf("no id set on resource %s", resourceName)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn

		_, err := conn.GetPrivateKey(&gofastly.GetPrivateKeyInput{
			ID: res.Primary.ID,
		})
		if err != nil {
			return fmt.Errorf("error getting private key from Fastly: %w", err)
		}

		return nil
	}
}
