package fastly

import (
	"fmt"
	"reflect"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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


func TestAccFastlyServiceV1_updateDomain(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	nameUpdate := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName1 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	domainName2 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config(name, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes(&service, name, []string{domainName1}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", name),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "1"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "1"),
				),
			},

			{
				Config: testAccServiceV1Config_domainUpdate(nameUpdate, domainName1, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes(&service, nameUpdate, []string{domainName1, domainName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "name", nameUpdate),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "2"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "domain.#", "2"),
				),
			},
		},
	})
}



func testAccCheckFastlyServiceV1Attributes(service *gofastly.ServiceDetail, name string, domains []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		domainList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Domains for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		expected := len(domains)
		for _, d := range domainList {
			for _, e := range domains {
				if d.Name == e {
					expected--
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Domain count mismatch, expected: %#v, got: %#v", domains, domainList)
		}

		return nil
	}
}

func testAccCheckFastlyServiceV1Attributes_backends(service *gofastly.ServiceDetail, name string, backends []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		expected := len(backendList)
		for _, b := range backendList {
			for _, e := range backends {
				if b.Address == e {
					expected--
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Backend count mismatch, expected: %#v, got: %#v", backends, backendList)
		}

		return nil
	}
}

func TestAccFastlyServiceV1_defaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	backendName2 := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backend(name, domain, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3600"),
				),
			},

			{
				Config: testAccServiceV1Config_backend_update(name, domain, backendName, backendName2, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3400"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "2"),
				),
			},
			// Now update the default_ttl to 0 and encounter the issue https://github.com/hashicorp/terraform/issues/12910
			{
				Config: testAccServiceV1Config_backend_update(name, domain, backendName, backendName2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName, backendName2}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "0"),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "active_version", "3"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_createDefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backendTTL(name, domain, backendName, 3400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3400"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_createZeroDefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domain := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_backendTTL(name, domain, backendName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "0"),
				),
			},
		},
	})
}


func testAccServiceV1Config(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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
}`, name, domain)
}

func testAccServiceV1Config_basicUpdate(name, comment, versionComment, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
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

func testAccServiceV1Config_domainUpdate(name, domain1, domain2 string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  domain {
    name    = "%s"
    comment = "tf-testing-other-domain"
  }

  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }

  force_destroy = true
}`, name, domain1, domain2)
}

func testAccServiceV1Config_backend(name, domain, backend string) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  force_destroy = true
}`, name, domain, backend)
}

func testAccServiceV1Config_backendTTL(name, domain, backend string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  default_ttl = %d

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  force_destroy = true
}`, name, ttl, domain, backend)
}

func testAccServiceV1Config_backend_update(name, domain, backend, backend2 string, ttl uint) string {
	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

	default_ttl = %d

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf-test-backend"
  }

  backend {
    address = "%s"
    name    = "tf-test-backend-other"
  }

  force_destroy = true
}`, name, ttl, domain, backend, backend2)
}
