package fastly

import (
	"errors"
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"sort"
	"strconv"
)

func dataSourceFastlyWAFRules() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyWAFRulesRead,

		Schema: map[string]*schema.Schema{
			"publishers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
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
						"latest_revision_number": {
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
	input := &gofastly.ListAllWAFRulesInput{}

	if v, ok := d.GetOk("publishers"); ok {
		l := v.([]interface{})
		for i := range l {
			input.FilterPublishers = append(input.FilterPublishers, l[i].(string))
		}
	}

	if v, ok := d.GetOk("tags"); ok {
		l := v.([]interface{})
		for i := range l {
			input.FilterTagNames = append(input.FilterTagNames, l[i].(string))
		}
	}

	log.Printf("[DEBUG] Reading WAF rules")
	res, err := conn.ListAllWAFRules(input)
	if err != nil {
		return fmt.Errorf("error listing WAF rules: %s", err)
	}

	d.SetId(strconv.Itoa(createFiltersHash(input)))

	rules := flattenWAFRules(res.Items)

	if err := d.Set("rules", rules); err != nil {
		return fmt.Errorf("error setting waf rules: %s", err)
	}

	return nil
}

func createFiltersHash(i *gofastly.ListAllWAFRulesInput) int {
	var result string
	for _, v := range i.FilterPublishers {
		result = result + v
	}
	for _, v := range i.FilterTagNames {
		result = result + v
	}
	return hashcode.String(result)
}

func flattenWAFRules(ruleList []*gofastly.WAFRule) []map[string]interface{} {

	var rl []map[string]interface{}
	if len(ruleList) == 0 {
		return rl
	}

	for _, r := range ruleList {

		var latestRevisionNumber int
		latestRevision, err := determineLatestRevision(r.Revisions)
		if err != nil {
			latestRevisionNumber = 1
		} else {
			latestRevisionNumber = latestRevision.Revision
		}

		rulesMapString := map[string]interface{}{
			"mod_sec_id":             r.ModSecID,
			"latest_revision_number": latestRevisionNumber,
			"type":                   r.Type,
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

func determineLatestRevision(versions []*gofastly.WAFRuleRevision) (*gofastly.WAFRuleRevision, error) {

	if len(versions) == 0 {
		return nil, errors.New("the list of WAFRuleRevisions cannot be empty")
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Revision > versions[j].Revision
	})

	return versions[0], nil
}
