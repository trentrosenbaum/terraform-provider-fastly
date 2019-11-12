package fastly

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var serviceRef = "fastly_service_v1.foo"
var condition = "prefetch"
var response = "response"

func TestResourceFastlyFlattenWAF(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAF
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.WAF{
				{
					ID:                "test1",
					PrefetchCondition: "prefetch",
					Response:          "response",
				},
			},
			local: []map[string]interface{}{
				{
					"waf_id":             "test1",
					"prefetch_condition": "prefetch",
					"response_object":    "response",
				},
			},
		},
	}
	for _, c := range cases {
		out := flattenWAFs(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceV1WAFAdd(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1(name, response, condition, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFAddAndRemove(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1(name, response, condition, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				Config: testAccServiceV1(name, response, condition, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
			{
				Config: testAccServiceV1(name, response, condition, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1DeletedWAF(&service),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFUpdateResponse(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updateResponse := "UpdatedResponse"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1(name, response, condition, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
			{
				Config: testAccServiceV1(name, updateResponse, condition, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, updateResponse, condition),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1WAFUpdateCondition(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	updatedCondition := "UpdatedPrefetch"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1(name, response, condition, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, condition),
				),
			},
			{
				Config: testAccServiceV1(name, response, updatedCondition, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceV1AttributesWAF(&service, name, response, updatedCondition),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceV1DeletedWAF(service *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafs, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			Service: service.ID,
			Version: strconv.Itoa(service.ActiveVersion.Number),
		})
		if err != nil {
			return err
		}

		if len(wafs) > 0 {
			return fmt.Errorf("[ERR] Error WAF %s should not be present for (%s), version (%v): %s", wafs[0].ID, service.ID, service.ActiveVersion.Number, err)
		}
		return nil
	}
}

func testAccCheckFastlyServiceV1AttributesWAF(service *gofastly.ServiceDetail, name, response, condition string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafs, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			Service: service.ID,
			Version: strconv.Itoa(service.ActiveVersion.Number),
		})

		waf, err := conn.GetWAF(&gofastly.GetWAFInput{
			Service: service.ID,
			Version: strconv.Itoa(service.ActiveVersion.Number),
			ID:      wafs[0].ID,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if waf.Response != response {
			return fmt.Errorf("WAF response mismatch, expected: %s, got: %#v", response, waf.Response)
		}

		if waf.PrefetchCondition != condition {
			return fmt.Errorf("WAF condition mismatch, expected: %#v, got: %#v", condition, waf.PrefetchCondition)
		}

		return nil
	}
}

func testAccServiceV1(name, response, condition string, withWAF bool) string {

	var waf string
	if withWAF {
		waf = fmt.Sprintf(`
		waf { 
			prefetch_condition = "%s" 
			response_object = "%s"
		}`, condition, response)
	}

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

  condition {
	name = "%s"
	type = "PREFETCH"
	statement = "req.url~+\"index.html\""
  }

  response_object {
	name = "%s"
	status = "403"
	response = "Forbidden"
	content = "content"
  }

  %s

  force_destroy = true
}`, name, domainName, backendName, condition, response, waf)

}
