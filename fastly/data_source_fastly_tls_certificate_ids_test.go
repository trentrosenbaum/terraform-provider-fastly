package fastly

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"testing"
)

func TestAccFastlyDataSourceTlSCertificateIDs(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyDataSourceTLSCertificateIDs,
				Check:  resource.TestCheckResourceAttrSet("data.fastly_tls_certificate_ids.subject", "ids.#"),
			},
		},
	})
}

const testAccFastlyDataSourceTLSCertificateIDs = `data "fastly_tls_certificate_ids" "subject" {}`
