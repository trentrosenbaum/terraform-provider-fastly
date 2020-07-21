package fastly

import (
	"fmt"
	"strconv"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// SuperTestBaseData provides a data structure for top level service attributes - i.e. those not in blocks.
type SuperTestBaseData struct {
	Name           string
	Comment        string
	VersionComment string
	Activate       bool
	ForceDestroy   bool
	DefaultTTL     int
	DefaultHost    string
}

// SuperTestComponentBase mixes-in the shared SuperTestComponent funcs with SuperTestBaseData.
type SuperTestComponentBase struct {
	*SuperTestComponentShared
	*SuperTestBaseData
}

// NewSuperTestComponentBase is the constructor for a new SuperTestComponentBase.
// It injects the service metadata and creates a random dataset for the base data.
func NewSuperTestComponentBase(service *SuperTestService) *SuperTestComponentBase {
	var c SuperTestComponentBase
	c.SuperTestComponentShared = &SuperTestComponentShared{
		service,
	}
	c.SuperTestBaseData = c.randomBase()
	return &c
}

// randomBase creates a random SuperTestBaseData.
func (c *SuperTestComponentBase) randomBase() *SuperTestBaseData {
	return &SuperTestBaseData{
		Name:           fmt.Sprintf("tf_test_%s", acctest.RandString(10)),
		Comment:        fmt.Sprintf("tf-test-%s", acctest.RandString(10)),
		VersionComment: "",
		Activate:       true,
		ForceDestroy:   true,
		DefaultTTL:     3600,
		DefaultHost:    "",
	}
}

// GetTestCheckFunc fulfils the SuperTestComponent interface.
// It composes testCheck functions for validating a resource that has been built.
func (c *SuperTestComponentBase) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc {
	testCheckFunc := []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "name", c.Name),
		resource.TestCheckResourceAttr(c.Service.ResourceId, "comment", c.Comment),
		resource.TestCheckResourceAttr(c.Service.ResourceId, "version_comment", c.VersionComment),
		resource.TestCheckResourceAttr(c.Service.ResourceId, "activate", strconv.FormatBool(c.Activate)),
		resource.TestCheckResourceAttr(c.Service.ResourceId, "force_destroy", strconv.FormatBool(c.ForceDestroy)),
	}
	if c.Service.Type == ServiceTypeVCL {
		testCheckFunc = append(testCheckFunc, []resource.TestCheckFunc{
			resource.TestCheckResourceAttr(c.Service.ResourceId, "default_ttl", strconv.Itoa(c.DefaultTTL)),
			resource.TestCheckResourceAttr(c.Service.ResourceId, "default_host", c.DefaultHost),
		}...)
	}
	return testCheckFunc
}

// AddRandom fulfils the SuperTestComponent interface.
// It normally adds an additional block, but since this is a single-block component, it is not supported.
func (c *SuperTestComponentBase) AddRandom() error {
	return fmt.Errorf("cannot add to base components")
}

// DeleteRandom fulfils the SuperTestComponent interface.
// It normally deletes a block, but since this is a single-block component, it is not supported.
func (c *SuperTestComponentBase) DeleteRandom() error {
	return fmt.Errorf("cannot delete from base components")
}

// UpdateRandom fulfils the SuperTestComponent interface.
// It updates the SuperTestBaseData to a new random configuration.
func (c *SuperTestComponentBase) UpdateRandom() error {
	c.SuperTestBaseData = c.randomBase()
	return nil
}
