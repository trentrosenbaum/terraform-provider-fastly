package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)


type FastlyTestServiceBlockAcl struct {
	Name  string
	AclID int
}

type FastlyTestServiceBlockAclResource struct {
	Name         string
	Domain   	 FastlyTestServiceBlockDomain
	BackendName  string
	ACLTest 	[]FastlyTestServiceBlockAcl
}

type TestResourceFastlyBuildAclCaseRemote struct {
	raw interface{}
	serviceID string
	serviceVersion int
}

func randomFastlyTestServiceBlockAcl() FastlyTestServiceBlockAcl {
	return FastlyTestServiceBlockAcl{
		acctest.RandString(12),
		acctest.RandInt(),
	}
}

func randomFastlyTestServiceBlockAcls(i int) []FastlyTestServiceBlockAcl {
	r := make([]FastlyTestServiceBlockAcl, i)
	for n := range r {
		r[n] = randomFastlyTestServiceBlockAcl()
	}
	return r
}


func TestResourceFastlyFlattenAcl(t *testing.T) {
	cases := []struct {
		sm     ServiceMetadata
		remote []*gofastly.ACL
		local  []map[string]interface{}
	}{
		{
			sm: ServiceMetadata{
				serviceType: ServiceTypeVCL,
			},
			remote: []*gofastly.ACL{
				{
					ID:   "1234567890",
					Name: "acl-example",
				},
			},
			local: []map[string]interface{}{
				{
					"acl_id": "1234567890",
					"name":   "acl-example",
				},
			},
		},
		{
			sm: ServiceMetadata{
				serviceType: ServiceTypeCompute,
			},
			remote: []*gofastly.ACL{
				{
					ID:   "1234567890",
					Name: "acl-example",
				},
			},
			local: []map[string]interface{}{
				{
					"acl_id": "1234567890",
					"name":   "acl-example",
				},
			},
		},
	}

	for _, c := range cases {
		acl := NewServiceACL(c.sm)
		out := acl.flatten(acl.listToGeneric(c.remote))
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}



func TestResourceFastlyBuildDeleteAcl(t *testing.T) {
	cases := []struct {
		sm     ServiceMetadata
		remote TestResourceFastlyBuildAclCaseRemote
		local  *gofastly.DeleteACLInput
	}{
		{
			// serviceType only necessary for constructor.
			// Not pertinent to test.
			sm: ServiceMetadata{
				serviceType: ServiceTypeVCL,
			},
			remote: TestResourceFastlyBuildAclCaseRemote{
					raw: map[string] interface{}{
						"name": 	"psgpisabiasgsfag",
						},
					serviceID:  	"1234235562",
					serviceVersion: 1,
					},
			local: &gofastly.DeleteACLInput{
				Service: "1234235562",
				Version: 1,
				Name: "psgpisabiasgsfag",
			},
		},

	}

	for _, c := range cases {
		acl  := NewServiceACL(c.sm)
		out := acl.buildDelete(c.remote.raw, c.remote.serviceID, c.remote.serviceVersion)
		if !reflect.DeepEqual(out.(*gofastly.DeleteACLInput), c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}

func TestResourceFastlyBuildCreateAcl(t *testing.T) {
	cases := []struct {
		sm     ServiceMetadata
		remote TestResourceFastlyBuildAclCaseRemote
		local  *gofastly.CreateACLInput
	}{
		{
			sm: ServiceMetadata{
				serviceType: ServiceTypeVCL,
			},
			remote: TestResourceFastlyBuildAclCaseRemote{
				raw: map[string] interface{}{
					"name": 	"wripiwrjhpasrgsar",
				},
				serviceID:  	"1234235562",
				serviceVersion: 1,
			},
			local: &gofastly.CreateACLInput{
				Service: "1234235562",
				Version: 1,
				Name: "wripiwrjhpasrgsar",
			},
		},

	}

	for _, c := range cases {
		acl  := NewServiceACL(c.sm)
		out := acl.buildCreate(c.remote.raw, c.remote.serviceID, c.remote.serviceVersion)
		if !reflect.DeepEqual(out.(*gofastly.CreateACLInput), c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}



func TestAccFastlyServiceV1_acl(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				PlanOnly: true,
				ExpectNonEmptyPlan: true,
				Config: testGetResourceTemplate("acl/basic",FastlyTestServiceBlockAclResource {
					makeTestBlockName(),
					randomFastlyTestServiceBlockDomain(),
					randomFastlyTestServiceBlockBackend(),
					randomFastlyTestServiceBlockAcls(1),
,					},

				},
			},
			{
				//PlanOnly: true,
				//ExpectNonEmptyPlan: true,
				ExpectError: regexp.MustCompile("config is invalid: Unsupported block type: Blocks of type \"acl\" are not expected here."),
				Config: testAccServiceV1Config_aclInvalidCompute(),
			},
		},
	})
}



/*
func TestAccFastlyServiceV1_acl(t *testing.T) {

	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	aclName := fmt.Sprintf("acl %s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1Config_acl(name, aclName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.foo", &service),
					testAccCheckFastlyServiceV1Attributes_acl(&service, name, aclName),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1Attributes_acl(service *gofastly.ServiceDetail, name, aclName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		acl, err := conn.GetACL(&gofastly.GetACLInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
			Name:    aclName,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up ACL records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if acl.Name != aclName {
			return fmt.Errorf("ACL logging endpoint name mismatch, expected: %s, got: %#v", aclName, acl.Name)
		}

		return nil
	}
}

func testAccServiceV1Config_acl(name, aclName string) string {
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

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

  acl {
	name       = "%s"
  }

  force_destroy = true
}`, name, domainName, backendName, aclName)
}
*/


func testAccServiceV1Config_acl01() string {


	name := acctest.RandString(12)
	aclName := acctest.RandString(12)
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
`, name, domainName, backendName, aclName)
}


func testAccServiceV1Config_aclInvalidCompute() string {
	name := acctest.RandString(12)
	aclName := acctest.RandString(12)
	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_compute" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
	}

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  acl {
	name = "%s"
  }

  force_destroy = true
}`, name, domainName, backendName, aclName)
}