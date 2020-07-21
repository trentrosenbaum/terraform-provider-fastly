package fastly

import (
	"bytes"
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/mitchellh/copystructure"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"text/template"
)

// SuperTest is the top level object for managing a Super Test.
// It contains service metadata, components and a lookup list for the components.
// The lookup list allows iteration of components without reflection.
// We opted not to use a list for components to provide a clearer data model for updating components and working
// with golang templates.
// This models a single Fastly service (V1 or compute)
type SuperTest struct {
	*SuperTestService
	*SuperTestComponentBase
	*SuperTestComponentBackend
	*SuperTestComponentDomain
	*SuperTestComponentHealthCheck
	*SuperTestComponentPackage
	*SuperTestComponentCloudFiles
	lookup []SuperTestComponent
}

// GetConfig returns the HCL resource configuration for the service resource to be super-tested.
func (superTest *SuperTest) GetConfig() string {
	var b bytes.Buffer
	err := superTest.getTemplate().ExecuteTemplate(&b, superTest.Service.TemplateName, superTest)
	if err != nil {
		log.Fatalf("Cannot execute template: %v", err)
	}
	return b.String()
}

// getTemplate loads all the golang templates for the super test.
func (superTest *SuperTest) getTemplate() *template.Template {
	var templateFiles []string
	var templateFolder = fmt.Sprintf("test_resource_templates/super_test/")
	filepath.Walk(templateFolder, func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".tmpl" {
			templateFiles = append(templateFiles, path)
		}
		return nil
	})
	var funcs = template.FuncMap{"StringsJoin": strings.Join}
	return template.Must(template.New("").Funcs(funcs).ParseFiles(templateFiles...))
}

// GetTestCheckFunc returns a composed list of test functions for the current service resource.
func (superTest *SuperTest) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) resource.TestCheckFunc {
	var testCheckFunc = []resource.TestCheckFunc{
		testAccCheckServiceV1Exists(superTest.Service.ResourceId, serviceDetail),
	}
	for _, c := range superTest.lookup {
		testCheckFunc = append(testCheckFunc, c.GetTestCheckFunc(serviceDetail)...)
	}

	return resource.ComposeTestCheckFunc(testCheckFunc...)
}

// GetConfig returns the HCL resource configuration for the service resource to be super-tested.
func (superTest *SuperTest) Copy() *SuperTest {
	var copyInterface = copystructure.Must(copystructure.Copy(superTest))
	var newSuperTest = copyInterface.(*SuperTest)
	newSuperTest.initLookup()
	return newSuperTest
}

// InitLookup uses a bit of reflection to configure the lookup list of components.
// This needs to be dynamic so that it works correctly on copy
func (superTest *SuperTest) initLookup() {
	var lookup []SuperTestComponent
	var reflect = reflect.ValueOf(*superTest)
	var prefix = "*fastly.SuperTestComponent"

	for i := 0; i < reflect.NumField(); i++ {
		f := reflect.Field(i)
		if !f.IsNil() && strings.HasPrefix(f.Type().String(), prefix) {
			lookup = append(lookup, f.Interface().(SuperTestComponent))
		}
	}

	superTest.lookup = lookup
}
