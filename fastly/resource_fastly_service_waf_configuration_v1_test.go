package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"testing"
)

func TestAccFastlyServiceWafConfigurationDetermineVersionV1(t *testing.T) {

	cases := []struct {
		remote []*gofastly.WAFVersion
		local  int
	}{
		{
			remote: []*gofastly.WAFVersion{
				{
					Number: 1,
				},
				{
					Number: 2,
				},
			},
			local: 2,
		},
		{
			remote: []*gofastly.WAFVersion{
				{
					Number: 3,
				},
				{
					Number: 2,
				},
				{
					Number: 1,
				},
			},
			local: 3,
		},
	}

	for _, c := range cases {
		out := determineVersion(c.remote)
		if c.local != out.Number {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestRefreshWAFVersion(t *testing.T) {

	var d schema.ResourceData
	var v gofastly.WAFVersion

	d.Set("allowed_http_versions", "anything")
	//
	refreshWAFVersion(&d, &v)

	ok, b := d.GetOk("allowed_http_versions")

	fmt.Printf("ok = %v b = %v \n", ok, b)

}
