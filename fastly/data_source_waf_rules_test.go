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

func TestAccFastlyWAFRulesDetermineRevision(t *testing.T) {

	cases := []struct {
		remote  []*gofastly.WAFRuleRevision
		local   int
		Errored bool
	}{
		{
			remote:  []*gofastly.WAFRuleRevision{},
			local:   0,
			Errored: true,
		},
		{
			remote: []*gofastly.WAFRuleRevision{
				{Revision: 1},
			},
			local:   1,
			Errored: false,
		},
		{
			remote: []*gofastly.WAFRuleRevision{
				{Revision: 1},
				{Revision: 2},
			},
			local:   2,
			Errored: false,
		},
		{
			remote: []*gofastly.WAFRuleRevision{
				{Revision: 3},
				{Revision: 2},
				{Revision: 1},
			},
			local:   3,
			Errored: false,
		},
	}

	for _, c := range cases {
		out, err := determineLatestRuleRevision(c.remote)
		if (err == nil) == c.Errored {
			t.Fatalf("Error expected to be %v but wan't", c.Errored)
		}
		if out == nil {
			continue
		}
		if c.local != out.Revision {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyWAFRulesFlattenWAFRules(t *testing.T) {
	cases := []struct {
		remote []*gofastly.WAFRule
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.WAFRule{
				{
					ModSecID: 11110000,
					Type:     "type",
					Revisions: []*gofastly.WAFRuleRevision{
						{Revision: 1},
					},
				},
			},
			local: []map[string]interface{}{
				{
					"modsec_rule_id":         11110000,
					"type":                   "type",
					"latest_revision_number": 1,
				},
			},
		},
	}
	for _, c := range cases {
		out := flattenWAFRules(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyWAFRulesPublisherFilter(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	wafrulesHCL := `
	publishers = ["owasp"]
    `
	wafrulesHCL2 := `
	publishers = ["owasp","fastly"]
    `
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(name, wafrulesHCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFRulesCheckByPublisherFilter(&service, 1, []string{"owasp"}),
				),
			},
			{
				Config: testAccFastlyWAFRules(name, wafrulesHCL2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFRulesCheckByPublisherFilter(&service, 2, []string{"owasp", "fastly"}),
				),
			},
		},
	})
}

func TestAccFastlyWAFRulesExcludeFilter(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	wafrulesHCL := `
	publishers = ["owasp"]
    exclude_modsec_rule_ids = [1010020]
    `
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(name, wafrulesHCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFRulesCheckByExcludeFilter(&service, 1, []string{"owasp"}, []int{1010020}),
				),
			},
		},
	})
}

func TestAccFastlyWAFRulesTagFilter(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	wafrulesHCL := `
	tags = ["CVE-2018-17384"]
    `

	wafrulesHCL2 := `
	tags = ["CVE-2018-17384", "attack-rce"]
    `
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyWAFRules(name, wafrulesHCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFRulesCheckByTagFilter(&service, 1, []string{"CVE-2018-17384"}),
				),
			},
			{
				Config: testAccFastlyWAFRules(name, wafrulesHCL2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFRulesCheckByTagFilter(&service, 2, []string{"CVE-2018-17384", "attack-rce"}),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceWAFRulesCheckByPublisherFilter(service *gofastly.ServiceDetail, wafVer int, publishers []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterPublishers: publishers,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		activeRules, err := testAccCheckFastlyServiceWAFRulesCheckWAFRules(service, wafVer)

		if len(rulesResp.Items) != len(activeRules) {
			return fmt.Errorf("[ERR] Expected waf rule size (%d), got (%d)", len(rulesResp.Items), len(activeRules))
		}
		return nil
	}
}

func testAccCheckFastlyServiceWAFRulesCheckByExcludeFilter(service *gofastly.ServiceDetail, wafVer int, publishers []string, exclude []int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterPublishers: publishers,
			ExcludeMocSecIDs: exclude,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		activeRules, err := testAccCheckFastlyServiceWAFRulesCheckWAFRules(service, wafVer)

		if len(rulesResp.Items) != len(activeRules) {
			return fmt.Errorf("[ERR] Expected waf rule size (%d), got (%d)", len(rulesResp.Items), len(activeRules))
		}
		return nil
	}
}

func testAccCheckFastlyServiceWAFRulesCheckByTagFilter(service *gofastly.ServiceDetail, wafVer int, tags []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		rulesResp, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
			FilterTagNames: tags,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		activeRules, err := testAccCheckFastlyServiceWAFRulesCheckWAFRules(service, wafVer)

		if len(rulesResp.Items) != len(activeRules) {
			return fmt.Errorf("[ERR] Expected waf rule size (%d), got (%d)", len(rulesResp.Items), len(activeRules))
		}
		return nil
	}
}

func testAccCheckFastlyServiceWAFRulesCheckWAFRules(service *gofastly.ServiceDetail, wafVer int) ([]*gofastly.WAFActiveRule, error) {
	conn := testAccProvider.Meta().(*FastlyClient).conn
	wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
		FilterService: service.ID,
		FilterVersion: service.ActiveVersion.Number,
	})
	if err != nil {
		return nil, fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
	}
	if len(wafResp.Items) != 1 {
		return nil, fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
	}

	waf := wafResp.Items[0]
	activeRulesResp, err := conn.ListAllWAFActiveRules(&gofastly.ListAllWAFActiveRulesInput{
		WAFID:            waf.ID,
		WAFVersionNumber: wafVer,
	})
	if err != nil {
		return nil, fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
	}
	return activeRulesResp.Items, nil
}

func testAccFastlyWAFRules(name, filtersHCL string) string {

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
	name = "prefetch"
	type = "PREFETCH"
	statement = "req.url~+\"index.html\""
  }
  response_object {
	name = "response"
	status = "403"
	response = "Forbidden"
	content = "content"
  }
  waf { 
	prefetch_condition = "prefetch" 
	response_object = "response"
  }
  force_destroy = true
}
variable "type_status" {
  type = map(string)
  default = {
    score     = "score"
    threshold = "log"
    strict    = "log"
  }
}
data "fastly_waf_rules" "r1" {
  %s
}
resource "fastly_service_waf_configuration_v1" "waf" {
  waf_id                          = fastly_service_v1.foo.waf[0].waf_id
  http_violation_score_threshold  = 202
  dynamic "rule" {
    for_each = data.fastly_waf_rules.r1.rules
    content {
      modsec_rule_id = rule.value.modsec_rule_id
      revision       = rule.value.latest_revision_number
      status         = lookup(var.type_status, rule.value.type, "log")
    }
  }
}
`, name, domainName, backendName, filtersHCL)
}
