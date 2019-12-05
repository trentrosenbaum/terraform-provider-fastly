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
				Description: "A list of publishers to be used filters for the data set.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of tags to be used as filters for the data set.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"exclude_modsec_rule_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A list of rules to be excluded from the data set referenced by modsecurity rule id.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"rules": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of rules that results from any given combination of filters.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"modsec_rule_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The modsecurity rule id.",
						},
						"latest_revision_number": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The modsecurity rule's latest revision.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The modsecurity rule's type.",
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

	if v, ok := d.GetOk("exclude_modsec_rule_ids"); ok {
		l := v.([]interface{})
		for i := range l {
			input.ExcludeMocSecIDs = append(input.ExcludeMocSecIDs, l[i].(int))
		}
	}

	log.Printf("[DEBUG] Reading WAF rules")
	res, err := conn.ListAllWAFRules(input)
	if err != nil {
		return fmt.Errorf("error listing WAF rules: %s", err)
	}

	rules := flattenWAFRules(res.Items)

	d.SetId(strconv.Itoa(createFiltersHash(input)))
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
	for _, v := range i.ExcludeMocSecIDs {
		result = result + strconv.Itoa(v)
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
		latestRevision, err := determineLatestRuleRevision(r.Revisions)
		if err != nil {
			latestRevisionNumber = 1
		} else {
			latestRevisionNumber = latestRevision.Revision
		}

		rulesMapString := map[string]interface{}{
			"modsec_rule_id":         r.ModSecID,
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

func determineLatestRuleRevision(versions []*gofastly.WAFRuleRevision) (*gofastly.WAFRuleRevision, error) {

	if len(versions) == 0 {
		return nil, errors.New("the list of WAFRuleRevisions cannot be empty")
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Revision > versions[j].Revision
	})

	return versions[0], nil
}
