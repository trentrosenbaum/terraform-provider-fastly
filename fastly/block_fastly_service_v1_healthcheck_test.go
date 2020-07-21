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

// SuperTestHealthCheckData provides a data structure for multiple HealthCheck blocks.
type SuperTestHealthCheckData struct {
	HealthChecks []*gofastly.HealthCheck
}

// SuperTestComponentHealthCheck mixes-in the shared SuperTestComponent functions with SuperTestHealthCheckData.
type SuperTestComponentHealthCheck struct {
	*SuperTestComponentShared
	*SuperTestHealthCheckData
}

// NewSuperTestComponentHealthCheck is the constructor for a SuperTestComponentHealthCheck.
// It injects the service metadata and creates a single random SuperTestHealthCheckBlock.
func NewSuperTestComponentHealthCheck(service *SuperTestService) *SuperTestComponentHealthCheck {
	var c SuperTestComponentHealthCheck
	c.SuperTestComponentShared = &SuperTestComponentShared{
		service,
	}
	c.SuperTestHealthCheckData = &SuperTestHealthCheckData{
		[]*gofastly.HealthCheck{c.randomHealthCheck()},
	}
	return &c
}

// randomHealthCheck creates a random SuperTestHealthCheckBlock.
func (c *SuperTestComponentHealthCheck) randomHealthCheck() *gofastly.HealthCheck {
	return &gofastly.HealthCheck{
		Name:             fmt.Sprintf("tf-test-healthcheck-%s", acctest.RandString(5)),
		Host:             fmt.Sprintf("%s.tf-test.fastly.com", acctest.RandString(5)),
		Path:             "/",
		CheckInterval:    5000,
		ExpectedResponse: 200,
		HTTPVersion:      "1.1",
		Initial:          2,
		Method:           "HEAD",
		Threshold:        3,
		Timeout:          500,
		Window:           5,
	}
}

// AddRandom fulfils the SuperTestComponent interface.
// It appends a random SuperTestHealthCheckBlock to the list.
func (c *SuperTestComponentHealthCheck) AddRandom() error {
	c.HealthChecks = append(c.HealthChecks, c.randomHealthCheck())
	return nil
}

// UpdateRandom fulfils the SuperTestComponent interface.
// It updates a randomly selected SuperTestHealthCheckBlock to a new random set of values.
func (c *SuperTestComponentHealthCheck) UpdateRandom() error {
	var randomIndex = c.randomIndex()
	c.HealthChecks[randomIndex] = c.randomHealthCheck()
	return nil
}

// DeleteRandom fulfils the SuperTestComponent interface.
// It deletes a random SuperTestHealthCheckBlock.
func (c *SuperTestComponentHealthCheck) DeleteRandom() error {
	var numDomains = len(c.HealthChecks)
	if numDomains < 2 {
		return fmt.Errorf("cannot delete random domain only have %v", numDomains)
	}
	var randomIndex = c.randomIndex()
	c.HealthChecks = append(c.HealthChecks[:randomIndex], c.HealthChecks[randomIndex+1:]...)
	return nil
}

// randomIndex selects an index at random from the list of SuperTestHealthCheckBlock.
func (c *SuperTestComponentHealthCheck) randomIndex() int {
	return rand.Intn(len(c.HealthChecks))
}

// GetTestCheckFunc fulfils the SuperTestComponent interface.
// It composes a series of tests to verify this component of a provisioned resource against
// the data used to provision it.
func (c *SuperTestComponentHealthCheck) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc {
	var numHealthChecks = len(c.HealthChecks)
	var r = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "healthcheck.#", strconv.Itoa(numHealthChecks)),
		c.testState(),
		c.testApi(serviceDetail),
	}
	return r
}

// testState checks a provisioned resource HealthCheck *state* against the data used to provision it.
func (c *SuperTestComponentHealthCheck) testState() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var sourceList = flattenHealthchecks(c.HealthChecks)
		var targetList, err = c.getStateTypeSetBlocks(s, "healthcheck")
		if err != nil {
			return err
		}

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("state healthcheck mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testApi checks a provisioned resource Backend at the *api* against the data used to provision it.
func (c *SuperTestComponentHealthCheck) testApi(serviceDetail *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		apiList, err := conn.ListHealthChecks(&gofastly.ListHealthChecksInput{
			Service: serviceDetail.ID,
			Version: serviceDetail.Version.Number,
		})
		if err != nil {
			return err
		}

		var sourceList = flattenHealthchecks(c.HealthChecks)
		var targetList = flattenHealthchecks(apiList)

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("api healthcheck mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testEquivalent compares two flattened data Domains
func (c *SuperTestComponentHealthCheck) testEquivalent(ms1 []map[string]interface{}, ms2 []map[string]interface{}) bool {
	expected := len(ms1)
	for _, m1 := range ms1 {
		for _, m2 := range ms2 {
			if c.compareKey(m1, m2, "name") &&
				c.compareKey(m1, m2, "host") &&
				c.compareKey(m1, m2, "path") &&
				c.compareKey(m1, m2, "check_interval") &&
				c.compareKey(m1, m2, "expected_response") &&
				c.compareKey(m1, m2, "http_version") &&
				c.compareKey(m1, m2, "initial") &&
				c.compareKey(m1, m2, "method") &&
				c.compareKey(m1, m2, "threshold") &&
				c.compareKey(m1, m2, "timeout") &&
				c.compareKey(m1, m2, "window") {
				expected--
			}
		}
	}
	return expected <= 0
}

func TestResourceFastlyFlattenHealthChecks(t *testing.T) {
	cases := []struct {
		remote []*gofastly.HealthCheck
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.HealthCheck{
				{
					Version:          1,
					Name:             "myhealthcheck",
					Host:             "example1.com",
					Path:             "/test1.txt",
					CheckInterval:    4000,
					ExpectedResponse: 200,
					HTTPVersion:      "1.1",
					Initial:          2,
					Method:           "HEAD",
					Threshold:        3,
					Timeout:          5000,
					Window:           5,
				},
			},
			local: []map[string]interface{}{
				{
					"name":              "myhealthcheck",
					"host":              "example1.com",
					"path":              "/test1.txt",
					"check_interval":    uint(4000),
					"expected_response": uint(200),
					"http_version":      "1.1",
					"initial":           uint(2),
					"method":            "HEAD",
					"threshold":         uint(3),
					"timeout":           uint(5000),
					"window":            uint(5),
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenHealthchecks(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n got: %#v", c.local, out)
		}
	}

}
