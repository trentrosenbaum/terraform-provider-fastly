package fastly

import (
	"fmt"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

type testServiceType struct {
	code string
	ref  string
}

var TestServiceTypeVCL = testServiceType{
	code: "vcl",
	ref:  "fastly_service_v1.foo",
}

var TestServiceTypeWasm = testServiceType{
	code: "wasm",
	ref:  "fastly_service_wasm.foo",
}

var TestServiceTypes = []testServiceType{TestServiceTypeVCL,TestServiceTypeWasm}


func TestAccFastlyServiceV1(t *testing.T) {
	var service gofastly.ServiceDetail
	name := makeTestServiceName()
	nameWasm := makeTestServiceName()

	testDestroy := func(*terraform.State) error {

		// reach out and DELETE the service
		conn := testAccProvider.Meta().(*FastlyClient).conn
		_, err := conn.DeactivateVersion(&gofastly.DeactivateVersionInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})
		if err != nil {
			return err
		}

		return conn.DeleteService(&gofastly.DeleteServiceInput{
			ID: service.ID,
		}) // Either err or nil
	}

	cases := map[string]map[string]interface{}{
		"vcl_basic": {
			"name":        name,
			"domain_name": makeTestDomainName(),
		},
		"vcl_basic_update": {
			"name":            name,
			"domain_name":     makeTestDomainName(),
			"comment":         makeTestServiceComment(),
			"version_comment": makeTestServiceComment(),
		},
		"vcl_versionless": {
			"service_name":         makeTestServiceName(),
			"dictionary_name":      makeTestBlockName("dictionary"),
			"acl_name":             makeTestBlockName("acl"),
			"dynamic_snippet_name": makeTestBlockName("dynamic_snippet"),
			"domain_name":          makeTestDomainName(),
		},
		"wasm_basic": {
			"name":        nameWasm,
			"domain_name": makeTestDomainName(),
		},
		"wasm_basic_update": {
			"name":            nameWasm,
			"domain_name":     makeTestDomainName(),
			"comment":         makeTestServiceComment(),
			"version_comment": makeTestServiceComment(),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigVCLServiceV1(cases["vcl_basic"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "name", cases["vcl_basic_update"]["name"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "comment", ManagedByTerraform),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "version_comment", ""),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "1"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "domain.#", "1"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "backend.#", "1"),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1(cases["vcl_basic_update"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "name", cases["vcl_basic_update"]["name"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "comment", cases["vcl_basic_update"]["comment"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "version_comment", cases["vcl_basic_update"]["version_comment"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "2"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "domain.#", "1"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "backend.#", "1"),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1(cases["vcl_basic"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testResourceConfigVCLServiceV1_Versionless(cases["vcl_versionless"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists("fastly_service_v1.service", &service),
					resource.TestCheckResourceAttr("fastly_service_acl_entries_v1.entries", "entry.#", "1"),
					resource.TestCheckResourceAttr("fastly_service_dictionary_items_v1.items", "items.%", "3"),
					resource.TestCheckResourceAttrSet("fastly_service_dynamic_snippet_content_v1.dyn_content", "content"),
				),
			},
			{
				Config: testResourceConfigWasmServiceV1(cases["wasm_basic"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeWasm.ref, &service),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "name", cases["wasm_basic_update"]["name"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "comment", ManagedByTerraform),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "version_comment", ""),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "active_version", "1"),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "domain.#", "1"),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "backend.#", "1"),
				),
			},

			{
				Config: testResourceConfigWasmServiceV1(cases["wasm_basic_update"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeWasm.ref, &service),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "name", cases["wasm_basic_update"]["name"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "comment", cases["wasm_basic_update"]["comment"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "version_comment", cases["wasm_basic_update"]["version_comment"].(string)),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "active_version", "2"),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "domain.#", "1"),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "backend.#", "1"),
				),
			},
			{
				Config: testResourceConfigWasmServiceV1(cases["wasm_basic"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeWasm.ref, &service),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testResourceConfigVCLServiceV1(data map[string]interface{}) string {
	return testGetResourceTemplate("service_vcl_basic", data)
}

func testResourceConfigVCLServiceV1_Versionless(data map[string]interface{}) string {
	return testGetResourceTemplate("service_vcl_versionless", data)
}

func testResourceConfigWasmServiceV1(data map[string]interface{}) string {
	return testGetResourceTemplate("service_wasm_basic", data)
}

// testAccCheckServiceV1Exists verifies that a service in the state exists at the fastly API
// It works equally well for VCL and WASM services
func testAccCheckServiceV1Exists(n string, service *gofastly.ServiceDetail) resource.TestCheckFunc {
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

// testAccCheckServiceV1Destroy verifies that a service in the state exists at the fastly API
// It works equally well for VCL and WASM services
func testAccCheckServiceV1Destroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {

		// Ignore resource if not a fastly service.
		if rs.Type != "fastly_service_v1" && rs.Type != "fastly_service_wasm_v1" {
			continue
		}

		// Get a list of services.
		conn := testAccProvider.Meta().(*FastlyClient).conn
		l, err := conn.ListServices(&gofastly.ListServicesInput{})
		if err != nil {
			return fmt.Errorf("[WARN] Error listing servcies when deleting Fastly Service (%s): %s", rs.Primary.ID, err)
		}

		// Fail if we find a service from the state that is in the list of active services.
		for _, s := range l {
			if s.ID == rs.Primary.ID {
				// service still found
				return fmt.Errorf("[WARN] Tried deleting Service (%s), but was still found", rs.Primary.ID)
			}
		}
	}
	return nil
}
