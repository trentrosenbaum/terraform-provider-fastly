package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
)

func dataSourceFastlyWAFRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyWAFRulesRead,

		Schema: map[string]*schema.Schema{
			"publishers": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mod_sec_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "",
						},
						"revision": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
					},
				},
			},
		},
	}
}

func dataSourceFastlyWAFRulesRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*FastlyClient).conn

	p := d.Get("publishers").([]interface{})

	publishers := make([]string, len(p), len(p))
	for i := range p {
		publishers[i] = p[i].(string)
	}

	log.Printf("[DEBUG] Reading WAF rules")
	res, err := conn.ListAllWAFRules(&gofastly.ListAllWAFRulesInput{
		FilterPublisher: publishers,
	})
	if err != nil {
		return fmt.Errorf("Error listing WAF rules: %s", err)
	}

	//	d.SetId(strconv.Itoa(hashcode.String(string(res.Items))))
	d.SetId("test2")

	rules := flattenWAFRules(res.Items)

	if err := d.Set("rules", rules); err != nil {
		return fmt.Errorf("Error setting waf rules: %s", err)
	}

	return nil
}

func flattenWAFRules(ruleList []*gofastly.WAFRule) []map[string]interface{} {

	var rl []map[string]interface{}
	if len(ruleList) == 0 {
		return rl
	}

	for _, r := range ruleList {
		rulesMapString := map[string]interface{}{
			"mod_sec_id": r.ModSecID,
			"revision":   r.Revision,
			"type":       r.Type,
		}
		// prune any empty values that come from the default string value in structs
		for k, v := range rulesMapString {
			if v == "" {
				delete(rulesMapString, k)
			}
		}
		rl = append(rl, rulesMapString)
	}

	return rl
}
