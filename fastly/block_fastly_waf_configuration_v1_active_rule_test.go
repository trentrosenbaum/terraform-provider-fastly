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

func TestAccFastlyServiceWAFVersionV1AddWithRules(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	rules := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "log",
			Revision: 1,
		},
	}
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	rulesTF := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules)
	wafVer := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1UpdateRules(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	rules1 := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "log",
			Revision: 1,
		},
	}
	rules2 := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "block",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "block",
			Revision: 1,
		},
	}
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	rulesTF1 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules1)
	wafVer1 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF1)

	rulesTF2 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules2)
	wafVer2 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF2)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules1, 1),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules2, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1DeleteRules(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))

	rules1 := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "log",
			Revision: 1,
		},
		{
			ModSecID: 2037405,
			Status:   "log",
			Revision: 1,
		},
	}
	rules2 := []gofastly.WAFActiveRule{
		{
			ModSecID: 2029718,
			Status:   "block",
			Revision: 1,
		},
	}
	wafVerInput := testAccFastlyServiceWAFVersionV1BuildConfig(20)
	rulesTF1 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules1)
	wafVer1 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF1)

	rulesTF2 := testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules2)
	wafVer2 := testAccFastlyServiceWAFVersionV1ComposeConfiguration(wafVerInput, rulesTF2)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules1, 1),
				),
			},
			{
				Config: testAccFastlyServiceWAFVersionV1(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckRules(&service, rules2, 2),
				),
			},
		},
	})
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
			WAFID:            waf.ID,
			WAFVersionNumber: wafVerNo,
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

func testAccCheckFastlyServiceWAFVersionV1ComposeWAFRules(rules []gofastly.WAFActiveRule) string {
	var result string
	for _, r := range rules {
		rule := fmt.Sprintf(`
          rule {
            modsec_rule_id = %d
            revision = %d
            status = "%s"
          }`, r.ModSecID, r.Revision, r.Status)
		result = result + rule
	}
	return result
}
