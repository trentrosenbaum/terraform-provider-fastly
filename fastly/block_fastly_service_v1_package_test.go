package fastly

import (
	"fmt"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// SuperTestPackageData provides a data structure to contain the single package block.
// We provide this parent struct to namespace the package in the data structure
type SuperTestPackageData struct {
	Package SuperTestPackageBlock
}

// SuperTestComponentPackage provides a data structure for a package block.
type SuperTestPackageBlock struct {
	Filename string
}

// SuperTestComponentPackage mixes-in the shared SuperTestComponent functions with SuperTestPackageData.
type SuperTestComponentPackage struct {
	*SuperTestComponentShared
	*SuperTestPackageData
	packageName string
}

// NewSuperTestComponentPackage is the constructor for a SuperTestComponentPackage.
// It injects the service metadata and creates a random SuperTestPackageBlock.
func NewSuperTestComponentPackage(packageName string, service *SuperTestService) *SuperTestComponentPackage {
	var c SuperTestComponentPackage
	c.SuperTestComponentShared = &SuperTestComponentShared{
		service,
	}
	c.SuperTestPackageData = &SuperTestPackageData{
		SuperTestPackageBlock{
			Filename: fmt.Sprintf("test_fixtures/package/%s.tar.gz", packageName),
		},
	}
	return &c
}

// AddRandom fulfils the SuperTestComponent interface.
// Not supported for Packages
func (c *SuperTestComponentPackage) AddRandom() error {
	return fmt.Errorf("cannot add to package component")
}

// UpdateRandom fulfils the SuperTestComponent interface.
// It updates a randomly selected SuperTestHealthCheckBlock to a new random set of values.
// ToDo: Allow update of packages
func (c *SuperTestComponentPackage) UpdateRandom() error {
	return nil
}

// DeleteRandom fulfils the SuperTestComponent interface.
// Not supported for Packages
func (c *SuperTestComponentPackage) DeleteRandom() error {
	return fmt.Errorf("cannot delete from package component")
}

// GetTestCheckFunc fulfils the SuperTestComponent interface.
// It composes a series of tests to verify this component of a provisioned resource against
// the data used to provision it.
func (c *SuperTestComponentPackage) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc {
	var r = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "package.#", "1"),
	}
	r = append(r, []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "package.0.filename", c.Package.Filename),
	}...)

	return r
}
