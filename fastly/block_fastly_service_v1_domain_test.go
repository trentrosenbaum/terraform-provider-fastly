package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestResourceFastlyFlattenDomains(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Domain
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Domain{
				{
					Name:    "test.notexample.com",
					Comment: "not comment",
				},
			},
			local: []map[string]interface{}{
				{
					"name":    "test.notexample.com",
					"comment": "not comment",
				},
			},
		},
		{
			remote: []*gofastly.Domain{
				{
					Name: "test.notexample.com",
				},
			},
			local: []map[string]interface{}{
				{
					"name":    "test.notexample.com",
					"comment": "",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDomains(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

type FastlyServiceV1DomainTestCase struct {
	Name    string
	Domains []FastlyServiceV1DomainTestCaseDomain
}

type FastlyServiceV1DomainTestCaseDomain struct {
	Name    string
	Comment string
}

func TestAccFastlyServiceV1_Domain(t *testing.T) {
	var service gofastly.ServiceDetail

	var domains = makeTestRandomDomainConfig(2)

	cases := map[string]FastlyServiceV1DomainTestCase{
		"vcl_domain": FastlyServiceV1DomainTestCase{
			Name:    makeTestServiceName(),
			Domains: domains[:1],
		},
		"vcl_domain_update": FastlyServiceV1DomainTestCase{
			Name:    makeTestServiceName(), // Update name
			Domains: domains[:2],
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigVCLServiceV1_Domains(cases["vcl_domain"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes(&service, cases["vcl_domain"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "name", cases["vcl_domain"].Name),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "active_version", "1"),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "domain.#", "1"),
				),
			},

			{
				Config: testResourceConfigVCLServiceV1_Domains(cases["vcl_domain_update"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes(&service, cases["vcl_domain_update"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "name", cases["vcl_domain_update"].Name),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "active_version", "2"),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "domain.#", "2"),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes(service *gofastly.ServiceDetail, data FastlyServiceV1DomainTestCase) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != data.Name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", data.Name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Domains for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		expected := len(data.Domains)
		for _, d := range domainList {
			for _, e := range data.Domains {
				if d.Name == e.Name {
					expected--
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Domain count mismatch, expected: %#v, got: %#v", data.Domains, domainList)
		}

		return nil
	}
}

func testResourceConfigVCLServiceV1_Domains(data FastlyServiceV1DomainTestCase) string {
	return testGetResourceTemplate("service_vcl_domains", data)
}

func makeTestRandomDomainConfig(num int) []FastlyServiceV1DomainTestCaseDomain {
	var domains = []FastlyServiceV1DomainTestCaseDomain{}
	for i := 0; i < num; i++ {
		domains = append(domains, FastlyServiceV1DomainTestCaseDomain{
			Name:    makeTestDomainName(),
			Comment: fmt.Sprintf("test-comment-%s", acctest.RandString(3)),
		})
	}
	return domains
}
