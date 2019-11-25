package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"sort"
	"testing"
)

func TestAccFastlyServiceWAFVersionV1DetermineVersion(t *testing.T) {

	cases := []struct {
		remote []*gofastly.WAFVersion
		local  int
	}{
		{
			remote: []*gofastly.WAFVersion{{Number: 1}, {Number: 2}},
			local:  2,
		},
		{
			remote: []*gofastly.WAFVersion{{Number: 3}, {Number: 2}, {Number: 1}},
			local:  3,
		},
	}

	for _, c := range cases {
		out := determineLatestVersion(c.remote)
		if c.local != out.Number {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceWAFVersionV1Add(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	threshold := 100
	wafVer := composeWAFConfiguration(threshold, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, threshold, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1AddWithRules(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	threshold := 100
	rules := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "log",
		},
		{
			ModSecID: 2037405,
			Status:   "log",
		},
	}
	rulesTF := composeWAFRule(rules)
	wafVer := composeWAFConfiguration(threshold, rulesTF)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1AddExistingService(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	threshold := 1001
	wafVer := composeWAFConfiguration(threshold, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, threshold, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Update(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	thresholdVersion1 := 1001
	wafVer1 := composeWAFConfiguration(thresholdVersion1, "")
	thresholdVersion2 := 1002
	wafVer2 := composeWAFConfiguration(thresholdVersion2, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, thresholdVersion1, 1),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, thresholdVersion2, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1AddUpdateRules(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	threshold := 100
	rules := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "log",
		},
		{
			ModSecID: 2037405,
			Status:   "log",
		},
	}
	rulesTF := composeWAFRule(rules)
	wafVer := composeWAFConfiguration(threshold, rulesTF)

	updatedRules := append(rules, gofastly.WAFActiveRule{
		ModSecID: 2074106,
		Status:   "block",
	})
	updatedRules[0].Status = "block"
	updatedRulesTF := composeWAFRule(updatedRules)
	updatedWAFVer := composeWAFConfiguration(threshold, updatedRulesTF)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules, 1),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, updatedWAFVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, updatedRules, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Delete(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	thresholdVersion1 := 1001
	wafVer1 := composeWAFConfiguration(thresholdVersion1, "")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, thresholdVersion1, 1),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, thresholdVersion1, 2),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceWAFVersionV1CheckAttributes(service *gofastly.ServiceDetail, threshold, latestVersion int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		verResp, err := conn.ListWAFVersions(&gofastly.ListWAFVersionsInput{
			WAFID: waf.ID,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF version records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(verResp.Items) < 1 {
			return fmt.Errorf("[ERR] Expected result size (%d), got (%d)", 1, len(verResp.Items))
		}

		wafVersion := verResp.Items[0]

		if threshold != wafVersion.InboundAnomalyScoreThreshold {
			return fmt.Errorf("WAF InboundAnomalyScoreThreshold mismatch, expected: %d, got: %d", threshold, wafVersion.InboundAnomalyScoreThreshold)
		}

		if threshold != wafVersion.LFIScoreThreshold {
			return fmt.Errorf("WAF LFIScoreThreshold mismatch, expected: %d, got: %d", threshold, wafVersion.LFIScoreThreshold)
		}

		if latestVersion != wafVersion.Number {
			return fmt.Errorf("WAF lastest vwrsion mismatch, expected: %d, got: %d", latestVersion, wafVersion.Number)
		}
		return nil
	}
}

func testAccCheckFastlyServiceWAFVersionV1CheckRules(service *gofastly.ServiceDetail, expected []gofastly.WAFActiveRule, wafVerNo int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		ruleResp, err := conn.ListWAFActiveRules(&gofastly.ListWAFActiveRulesInput{
			WAFID:     waf.ID,
			WAFNumber: wafVerNo,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		actual := ruleResp.Items
		if len(expected) != len(actual) {
			return fmt.Errorf("Error matching rules slice sizes :\nexpected: %#v\ngot: %#v", len(expected), len(actual))
		}

		sort.Slice(expected[:], func(i, j int) bool {
			return expected[i].ModSecID < expected[j].ModSecID
		})
		sort.Slice(actual[:], func(i, j int) bool {
			return actual[i].ModSecID < actual[j].ModSecID
		})
		for i := range expected {
			if expected[i].ModSecID != actual[i].ModSecID {
				return fmt.Errorf("Error matching:\nexpected: %#v\ngot: %#v", expected[i].ModSecID, actual[i].ModSecID)
			}
			if expected[i].Status != actual[i].Status {
				return fmt.Errorf("Error matching:\nexpected: %#v\ngot: %#v", expected[i].Status, actual[i].Status)
			}
		}
		return nil
	}
}

func composeWAFConfiguration(threshold int, rules string) string {
	return fmt.Sprintf(`
		resource "fastly_service_waf_configuration_v1" "waf" {
  			waf_id = fastly_service_v1.foo.waf[0].waf_id
  			inbound_anomaly_score_threshold = %d
  			lfi_score_threshold = %d
            %s
}`, threshold, threshold, rules)
}

func composeWAFRule(rules []gofastly.WAFActiveRule) string {
	var result string
	for _, r := range rules {
		rule := fmt.Sprintf(`
		  rule {
    		modsec_rule_id = %d
    		status = "%s"
  		}`, r.ModSecID, r.Status)
		result = result + rule
	}
	return result
}

func testAccServiceV1WAFVersion(name, extraHCL string) string {

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
  %s
`, name, domainName, backendName, extraHCL)
}
