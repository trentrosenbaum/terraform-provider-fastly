package fastly

import (
	"fmt"
	"testing"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/stretchr/testify/require"
)

func TestAccFastlyTLSCertificate_withName(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	updatedName := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example.com", name)

	key, cert, cert2, err := generateKeyAndMultipleCerts(domain)
	require.NoError(t, err)

	resourceName := "fastly_tls_certificate.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTLSCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTlsCertificateWithName(name, key, name, cert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "issued_to", domain),
					resource.TestCheckResourceAttrSet(resourceName, "issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "replace"),
					resource.TestCheckResourceAttrSet(resourceName, "serial_number"),
					resource.TestCheckResourceAttrSet(resourceName, "signature_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					testAccTLSConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccTlsCertificateWithName(name, key, updatedName, cert2),
				Check:  resource.TestCheckResourceAttr(resourceName, "name", updatedName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_body"},
			},
		},
	})
}

func TestAccFastlyTLSCertificate_withoutName(t *testing.T) {
	name := acctest.RandomWithPrefix(testResourcePrefix)
	domain := fmt.Sprintf("%s.example.com", name)

	key, cert, err := generateKeyAndCert(domain)
	require.NoError(t, err)

	resourceName := "fastly_tls_certificate.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckTLSCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTlsCertificateWithoutName(name, key, cert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", domain),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttrSet(resourceName, "updated_at"),
					resource.TestCheckResourceAttr(resourceName, "issued_to", domain),
					resource.TestCheckResourceAttrSet(resourceName, "issuer"),
					resource.TestCheckResourceAttrSet(resourceName, "replace"),
					resource.TestCheckResourceAttrSet(resourceName, "serial_number"),
					resource.TestCheckResourceAttrSet(resourceName, "signature_algorithm"),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					testAccTLSConfigurationExists(resourceName),
				),
			},
		},
	})
}

func testAccTLSConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		r := s.RootModule().Resources[resourceName]

		conn := testAccProvider.Meta().(*FastlyClient).conn

		_, err := conn.GetCustomTLSCertificate(&fastly.GetCustomTLSCertificateInput{
			ID: r.Primary.ID,
		})
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccTlsCertificateWithName(keyName string, key string, certName string, cert string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

resource "fastly_tls_certificate" "test" {
  name = "%[3]s"
  certificate_body = <<EOF
%[4]s
EOF
  depends_on = [fastly_tls_private_key.key]
}
`, keyName, key, certName, cert)
}

func testAccTlsCertificateWithoutName(keyName string, key string, cert string) string {
	return fmt.Sprintf(`
resource "fastly_tls_private_key" "key" {
  name = "%[1]s"
  key_pem = <<EOF
%[2]s
EOF
}

resource "fastly_tls_certificate" "test" {
  certificate_body = <<EOF
%[3]s
EOF
  depends_on = [fastly_tls_private_key.key]
}
`, keyName, key, cert)
}

func testAccCheckTLSCertificateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*FastlyClient).conn

	for _, r := range s.RootModule().Resources {
		if r.Type != "fastly_tls_certificate" {
			continue
		}

		certificates, err := conn.ListCustomTLSCertificates(&fastly.ListCustomTLSCertificatesInput{})
		if err != nil {
			return err
		}

		for _, certificate := range certificates {
			if certificate.ID == r.Primary.ID {
				return fmt.Errorf("certificate %s still exists", r.Primary.ID)
			}
		}
	}

	return nil
}