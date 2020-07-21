package fastly

import (
	"fmt"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// serviceMetadataV1 defines metadata for a vcl service.
var serviceMetadataV1 = SuperTestService{
	SuperTestServiceMetadata{
		Type:         ServiceTypeVCL,
		ResourceId:   "fastly_service_v1.foo",
		TemplateName: "service_v1",
	},
}

// NewV1SuperTest is a constructor for a vcl service SuperTest.
func NewV1SuperTest() *SuperTest {
	superTest := &SuperTest{
		SuperTestService:              &serviceMetadataV1,
		SuperTestComponentBase:        NewSuperTestComponentBase(&serviceMetadataV1),
		SuperTestComponentBackend:     NewSuperTestComponentBackend(&serviceMetadataV1),
		SuperTestComponentDomain:      NewSuperTestComponentDomain(&serviceMetadataV1),
		SuperTestComponentHealthCheck: NewSuperTestComponentHealthCheck(&serviceMetadataV1),
		SuperTestComponentCloudFiles:  NewSuperTestComponentCloudFiles(&serviceMetadataV1),
		SuperTestComponentPackage:     nil,
		lookup:                        nil, // See below
	}
	// Lookup provides an iterable list of components - using reflection in a defined location to initialise this list.
	// SuperTestComponent* objects are added, nil objects are ignored, allowing us to exclude irrelevant blocks,
	// such as Package for VCL services.
	superTest.initLookup()
	return superTest
}

func TestAccFastlyServiceV1(t *testing.T) {
	var serviceDetail gofastly.ServiceDetail

	// "Create resource" test
	test := NewV1SuperTest()

	// "Update blocks" test
	testUpdate := test.Copy()
	testUpdate.SuperTestComponentBackend.UpdateRandom()
	testUpdate.SuperTestComponentDomain.UpdateRandom()
	testUpdate.SuperTestComponentHealthCheck.UpdateRandom()

	// "Add blocks" test
	testAdd := testUpdate.Copy()
	testAdd.SuperTestComponentBackend.AddRandom()
	testAdd.SuperTestComponentDomain.AddRandom()
	testAdd.SuperTestComponentHealthCheck.AddRandom()

	// "Delete blocks" test
	testDelete := testAdd.Copy()
	testDelete.SuperTestComponentBackend.DeleteRandom()
	testDelete.SuperTestComponentDomain.DeleteRandom()
	testDelete.SuperTestComponentHealthCheck.DeleteRandom()

	// "Bad Backend" test
	testBadBackend := testDelete.Copy()
	testBadBackend.SuperTestComponentBackend.AddRandom()
	testBadBackend.Backends[len(testBadBackend.Backends)-1].Address = fmt.Sprintf("%s.aws.amazon.com.", acctest.RandString(3))

	// "Zero TTL"
	testZeroTTL := testBadBackend.Copy()
	testZeroTTL.Backends = testZeroTTL.Backends[:len(testZeroTTL.Backends)-1]
	testZeroTTL.DefaultTTL = 0

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: test.GetConfig(),
				Check:  test.GetTestCheckFunc(&serviceDetail),
			},
			{
				Config: testUpdate.GetConfig(),
				Check:  testUpdate.GetTestCheckFunc(&serviceDetail),
			},
			{
				Config: testAdd.GetConfig(),
				Check:  testAdd.GetTestCheckFunc(&serviceDetail),
			},
			{
				Config: testDelete.GetConfig(),
				Check:  testDelete.GetTestCheckFunc(&serviceDetail),
			},
			{
				Config:      testBadBackend.GetConfig(),
				ExpectError: regexp.MustCompile("Bad Request"),
			},
			// Now update the default_ttl to 0 and encounter the issue https://github.com/hashicorp/terraform/issues/12910
			{
				Config: testZeroTTL.GetConfig(),
				Check:  testZeroTTL.GetTestCheckFunc(&serviceDetail),
			},
		},
	})
}

// ServiceV1_disappears – test that a non-empty plan is returned when a Fastly
// Service is destroyed outside of Terraform, and can no longer be found,
// correctly clearing the ID field and generating a new plan
func TestAccFastlyServiceV1_disappears(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	testDestroy := func(*terraform.State) error {
		// reach out and DELETE the service
		conn := testAccProvider.Meta().(*FastlyClient).conn
		// deactivate active version to destoy
		_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})
		if err != nil {
			return err
		}

		// delete service
		err = conn.DeleteService(&gofastly.DeleteServiceInput{
			ID: service.ID,
		})

		if err != nil {
			return err
		}

		return nil
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config(name, domainName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccFastlyServiceV1_defaultTTL(t *testing.T) {
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
				Config: testAccServiceV1Config_backend(name, domain, backendName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_backends(&service, name, []string{backendName}),
					resource.TestCheckResourceAttr(
						"fastly_service_v1.foo", "default_ttl", "3600"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_import(t *testing.T) {
	var service gofastly.ServiceDetail
	serviceName := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1_import(serviceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
				),
			},
			{
				ResourceName:      "fastly_service_v1.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy"},
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

func testAccServiceV1_import(serviceName string) string {
	backendName01 := fmt.Sprintf("%s.%s.com", acctest.RandString(3), acctest.RandString(12))
	domainName01 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
	backendName02 := fmt.Sprintf("%s.%s.com", acctest.RandString(3), acctest.RandString(12))
	domainName02 := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain01"
	}
  domain {
    name    = "%s"
    comment = "tf-testing-domain02"
	}

  backend {
    address = "%s"
    name    = "tf -test backend 01"
  }

  backend {
    address = "%s"
    name    = "tf -test backend 02"
  }

  force_destroy = true
}

`, serviceName, domainName01, backendName01, domainName02, backendName02)
}
