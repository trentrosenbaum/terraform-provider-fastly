package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

var activeRule = &schema.Schema{
	Type:     schema.TypeSet,
	Optional: true,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"status": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "The Web Application Firewall rule's status. Allowed values are (log, block and score)",
				ValidateFunc: validateRuleStatusType(),
			},
			"modsec_rule_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The Web Application Firewall rule's modsec id",
			},
			"revision": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The Web Application Firewall rule's revision",
			},
		},
	},
}

func updateRules(d *schema.ResourceData, meta interface{}, wafID string, Number int) error {

	conn := meta.(*FastlyClient).conn
	os, ns := d.GetChange("rule")

	if os == nil {
		os = new(schema.Set)
	}
	if ns == nil {
		ns = new(schema.Set)
	}

	oss := os.(*schema.Set)
	nss := ns.(*schema.Set)

	remove := oss.Difference(nss).List()
	add := nss.Difference(oss).List()

	if len(remove) > 0 {
		deleteOpts := buildDeleteWAFRulesInput(remove, wafID, Number)
		log.Printf("[DEBUG] WAF rules delete opts: %#v", deleteOpts)
		err := conn.DeleteWAFActiveRules(&deleteOpts)
		if err != nil {
			return err
		}
	}

	if len(add) > 0 {
		createOpts := buildCreateWAFRulesInput(add, wafID, Number)
		log.Printf("[DEBUG] WAF rules create opts: %#v", createOpts)
		_, err := conn.CreateWAFActiveRules(&createOpts)
		if err != nil {
			return err
		}
	}

	return nil
}

func readWAFRules(meta interface{}, d *schema.ResourceData, v int) error {

	conn := meta.(*FastlyClient).conn
	wafID := d.Get("waf_id").(string)

	resp, err := conn.ListAllWAFActiveRules(&gofastly.ListAllWAFActiveRulesInput{
		WAFID:            wafID,
		WAFVersionNumber: v,
	})
	if err != nil {
		return err
	}

	rules := flattenWAFRules(resp.Items)
	if err := d.Set("rule", rules); err != nil {
		log.Printf("[WARN] Error setting WAF rules for (%s): %s", d.Id(), err)
	}
	return nil
}

func buildDeleteWAFRulesInput(add []interface{}, wafID string, wafVersionNumber int) gofastly.DeleteWAFActiveRulesInput {

	var rules []*gofastly.WAFActiveRule
	for _, rRaw := range add {
		rf := rRaw.(map[string]interface{})

		rules = append(rules, &gofastly.WAFActiveRule{
			ModSecID: rf["modsec_rule_id"].(int),
		})
	}

	return gofastly.DeleteWAFActiveRulesInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Rules:            rules,
	}
}

func buildCreateWAFRulesInput(add []interface{}, wafID string, wafVersionNumber int) gofastly.CreateWAFActiveRulesInput {

	var rules []*gofastly.WAFActiveRule
	for _, rRaw := range add {
		rf := rRaw.(map[string]interface{})

		rules = append(rules, &gofastly.WAFActiveRule{
			ModSecID: rf["modsec_rule_id"].(int),
			Revision: rf["revision"].(int),
			Status:   rf["status"].(string),
		})
	}

	return gofastly.CreateWAFActiveRulesInput{
		WAFID:            wafID,
		WAFVersionNumber: wafVersionNumber,
		Rules:            rules,
	}
}

func flattenWAFRules(rules []*gofastly.WAFActiveRule) []map[string]interface{} {
	var rl []map[string]interface{}
	for _, r := range rules {

		ruleMapString := map[string]interface{}{
			"modsec_rule_id": r.ModSecID,
			"revision":       r.Revision,
			"status":         r.Status,
		}

		rl = append(rl, ruleMapString)
	}
	return rl
}
