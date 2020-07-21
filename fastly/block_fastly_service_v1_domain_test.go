package fastly

import (
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// SuperTestDomainData provides a data structure for multiple domain blocks.
type SuperTestDomainData struct {
	Domains []*gofastly.Domain
}

// SuperTestComponentDomain mixes-in the shared SuperTestComponent functions with SuperTestDomainData.
type SuperTestComponentDomain struct {
	*SuperTestComponentShared
	*SuperTestDomainData
}

// NewSuperTestComponentDomain is the constructor for a SuperTestComponentDomain.
// It injects the service metadata and creates a single random SuperTestDomainBlock.
func NewSuperTestComponentDomain(service *SuperTestService) *SuperTestComponentDomain {
	var c SuperTestComponentDomain
	c.SuperTestComponentShared = &SuperTestComponentShared{
		service,
	}
	c.SuperTestDomainData = &SuperTestDomainData{
		[]*gofastly.Domain{c.randomDomain()},
	}
	return &c
}

// randomDomain creates a random SuperTestDomainBlock.
func (c *SuperTestComponentDomain) randomDomain() *gofastly.Domain {
	return &gofastly.Domain{
		Name:    fmt.Sprintf("%s.tf-test.fastly-%s.com", acctest.RandString(5), acctest.RandString(5)),
		Comment: fmt.Sprintf("tf-test-domain-%s", acctest.RandString(5)),
	}
}

// AddRandom fulfils the SuperTestComponent interface.
// It appends a random SuperTestDomainBlock to the list.
func (c *SuperTestComponentDomain) AddRandom() error {
	c.Domains = append(c.Domains, c.randomDomain())
	return nil
}

// UpdateRandom fulfils the SuperTestComponent interface.
// It updates a randomly selected SuperTestDomainBlock to a new random set of values.
func (c *SuperTestComponentDomain) UpdateRandom() error {
	var randomIndex = c.randomIndex()
	c.Domains[randomIndex] = c.randomDomain()
	return nil
}

// DeleteRandom fulfils the SuperTestComponent interface.
// It deletes a random SuperTestDomainBlock.
func (c *SuperTestComponentDomain) DeleteRandom() error {
	var numDomains = len(c.Domains)
	if numDomains < 2 {
		return fmt.Errorf("cannot delete random domain only have %v", numDomains)
	}
	var randomIndex = c.randomIndex()
	c.Domains = append(c.Domains[:randomIndex], c.Domains[randomIndex+1:]...)
	return nil
}

// randomIndex selects an index at random from the list of SuperTestDomainBlocks.
func (c *SuperTestComponentDomain) randomIndex() int {
	return rand.Intn(len(c.Domains))
}

// GetTestCheckFunc fulfils the SuperTestComponent interface.
// It composes a series of tests to verify this component of a provisioned resource against
// the data used to provision it.
func (c *SuperTestComponentDomain) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc {
	var numDomains = len(c.Domains)
	var r = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "domain.#", strconv.Itoa(numDomains)),
		c.testState(),
		c.testApi(serviceDetail),
	}
	return r
}

// testState checks a provisioned resource Domain *state* against the data used to provision it.
func (c *SuperTestComponentDomain) testState() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var sourceList = flattenDomains(c.Domains)
		var targetList, err = c.getStateTypeSetBlocks(s, "domain")
		if err != nil {
			return err
		}

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("state domain mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testApi checks a provisioned resource Domain at the *api* against the data used to provision it.
func (c *SuperTestComponentDomain) testApi(serviceDetail *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		apiList, err := conn.ListDomains(&gofastly.ListDomainsInput{
			Service: serviceDetail.ID,
			Version: serviceDetail.Version.Number,
		})
		if err != nil {
			return err
		}

		var sourceList = flattenDomains(c.Domains)
		var targetList = flattenDomains(apiList)

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("api domain mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testEquivalent compares two flattened data Domains
func (c *SuperTestComponentDomain) testEquivalent(ms1 []map[string]interface{}, ms2 []map[string]interface{}) bool {
	expected := len(ms1)
	for _, m1 := range ms1 {
		for _, m2 := range ms2 {
			if c.compareKey(m1, m2, "name") &&
				c.compareKey(m1, m2, "comment") {
				expected--
			}
		}
	}
	return expected <= 0
}

func TestResourceFastlyFlattenDomains(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Domain
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Domain{
				{
					Name:    "test.notexample.com",
					Comment: "not comment",
				},
			},
			local: []map[string]interface{}{
				{
					"name":    "test.notexample.com",
					"comment": "not comment",
				},
			},
		},
		{
			remote: []*gofastly.Domain{
				{
					Name: "test.notexample.com",
				},
			},
			local: []map[string]interface{}{
				{
					"name":    "test.notexample.com",
					"comment": "",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenDomains(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\ngot: %#v", c.local, out)
		}
	}
}
