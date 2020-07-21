package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// serviceMetadataCompute defines metadata for a compute service.
var serviceMetadataCompute = SuperTestService{
	SuperTestServiceMetadata{
		Type:         ServiceTypeCompute,
		ResourceId:   "fastly_service_compute.foo",
		TemplateName: "service_compute",
	},
}

// NewComputeSuperTest is a constructor for a compute service SuperTest.
func NewComputeSuperTest() *SuperTest {
	superTest := &SuperTest{
		SuperTestService:              &serviceMetadataCompute,
		SuperTestComponentBase:        NewSuperTestComponentBase(&serviceMetadataCompute),
		SuperTestComponentBackend:     NewSuperTestComponentBackend(&serviceMetadataCompute),
		SuperTestComponentDomain:      NewSuperTestComponentDomain(&serviceMetadataCompute),
		SuperTestComponentHealthCheck: NewSuperTestComponentHealthCheck(&serviceMetadataCompute),
		SuperTestComponentPackage:     NewSuperTestComponentPackage("valid", &serviceMetadataCompute),
		SuperTestComponentCloudFiles:  NewSuperTestComponentCloudFiles(&serviceMetadataCompute),
		lookup:                        nil, // See below
	}
	// Lookup provides an iterable list of components - using reflection in a defined location to initialise this list.
	// SuperTestComponent* objects are added, nil objects are ignored, allowing us to exclude irrelevant blocks,
	// such as VCL for compute services.
	superTest.initLookup()
	return superTest
}

func TestAccFastlyServiceCompute(t *testing.T) {
	var serviceDetail gofastly.ServiceDetail

	// "Create resource" test
	test := NewComputeSuperTest()

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
		},
	})
}

func testAccCheckServiceComputeV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "fastly_service_compute" {
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

func TestAccFastlyServiceComputeV1_import(t *testing.T) {
	var service gofastly.ServiceDetail
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceComputeV1ImportConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_compute.foo", &service),
				),
			},
			{
				ResourceName:      "fastly_service_compute.foo",
				ImportState:       true,
				ImportStateVerify: true,
				// These attributes are not stored on the Fastly API and must be ignored.
				ImportStateVerifyIgnore: []string{"activate", "force_destroy", "package.0.filename"},
			},
		},
	})

}

func testAccServiceComputeV1Config(name, domain string) string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"
  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }
  backend {
    address = "aws.amazon.com"
    name    = "amazon docs"
  }
  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }
  force_destroy = true
  activate = false
}`, name, domain)
}

func testAccServiceComputeV1ImportConfig() string {
	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "tf-test-%s"
  domain {
    name    = "fastly-import-test-%s.tf-fastly.com"
    comment = "fastly-import-test-domain-01"
  }
  domain {
    name    = "fastly-import-test-%s.tf-fastly.com"
    comment = "fastly-import-test-domain-02"
  }
  backend {
    address = "%s.%s.com"
    name    = "import test backend 01"
  }
 backend {
    address = "%s.%s.com"
    name    = "import test backend 02"
  }
  package {
    filename = "test_fixtures/package/valid.tar.gz"
	source_code_hash = filesha512("test_fixtures/package/valid.tar.gz")
  }
  force_destroy = true
  activate = false
}`, acctest.RandString(10), // name
		acctest.RandString(5),                        // domain01
		acctest.RandString(5),                        // domain02
		acctest.RandString(8), acctest.RandString(8), // backend01
		acctest.RandString(8), acctest.RandString(8), // backend02
	)
}
