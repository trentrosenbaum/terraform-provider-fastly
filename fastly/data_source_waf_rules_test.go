package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"reflect"
	"testing"
)

func TestAccFastlyServiceWAFRuleV1DetermineRevision(t *testing.T) {

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

func TestResourceFastlyFlattenWAFRules(t *testing.T) {
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
