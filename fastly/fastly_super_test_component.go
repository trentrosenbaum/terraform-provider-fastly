package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

// SuperTestComponent defines the common interface for all SuperTest components.
type SuperTestComponent interface {
	GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc
	AddRandom() error
	UpdateRandom() error
	DeleteRandom() error
}

// SuperTestComponentShared is a mixin to provide shared functions to SuperTest components.
type SuperTestComponentShared struct {
	*SuperTestService
}

// getStateTypeSetBlocks arranges block attributes in a TypeSet into an array rather than "block.462472.attribute"
func (h *SuperTestComponentShared) getStateTypeSetBlocks(s *terraform.State, blockType string) ([]map[string]interface{}, error) {

	// Get top level resource (e.g. fastly_service_v1.foo)
	rs, ok := s.RootModule().Resources[h.Service.ResourceId]
	if !ok {
		return nil, fmt.Errorf("not found: %s in state root", h.Service.ResourceId)
	}

	// Get primary
	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("no primary instance: %s in state root", h.Service.ResourceId)
	}

	var attr = map[string]map[string]interface{}{}
	for k, v := range is.Attributes {

		// Split the attribute key on "."
		// Should be [blockType].[blockId].[attribute]
		// For example, "backend.573578.name"
		var attrSplit = strings.Split(k, ".")
		if attrSplit[0] == blockType {
			// Ignore blockId # since this is metadata stating how many blocks we have
			if attrSplit[1] != "#" {
				if len(attrSplit) != 3 {
					return nil, fmt.Errorf("block %s has wrong number of elements (should be 3): %s", blockType, k)
				}
				// Ensure we have a map for this block id
				if _, ok := attr[attrSplit[1]]; !ok {
					attr[attrSplit[1]] = map[string]interface{}{}
				}
				attr[attrSplit[1]][attrSplit[2]] = v
			}
		}
	}

	// Convert map to array
	// We used a map previously to simplify grouping by block Id
	var r []map[string]interface{}
	for _, v := range attr {
		r = append(r, v)
	}

	return r, nil
}

// Get main serviceMetadata object (from test service metadata)
func (h *SuperTestComponentShared) getServiceMetadata() *ServiceMetadata {
	return &ServiceMetadata{
		serviceType: h.Service.Type,
	}
}

func (h *SuperTestComponentShared) compareKey(m1 map[string]interface{}, m2 map[string]interface{}, k string) bool {
	var i1 = m1[k]
	var i2 = m2[k]
	var ok bool

	if i1 == nil && i2 == nil {
		return true
	} else if i1 == nil {
		fmt.Println("nil value for i1 in compareKey")
		return false
	} else if i2 == nil {
		fmt.Println("nil value for i2 in compareKey")
		return false
	}

	kind1 := reflect.TypeOf(i1).Kind()
	kind2 := reflect.TypeOf(i2).Kind()

	if kind1 != kind2 {
		i1, ok = h.interfaceToString(i1)
		if !ok {
			return false
		}
		i2, ok = h.interfaceToString(i2)
		if !ok {
			return false
		}
	}
	return reflect.DeepEqual(i1, i2)
}

func (h *SuperTestComponentShared) interfaceToString(i interface{}) (string, bool) {
	switch reflect.TypeOf(i).Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		r := fmt.Sprintf("%v", i)
		return r, true
	}
	fmt.Printf("Cannot cast type %v to string\n", reflect.TypeOf(i).Kind())
	return "", false
}

// randomVariance returns a number plus or minus a number within the maxVariance
// Careful, this isn't closely type checked between int<-->uint so avoid large numbers
func (h *SuperTestComponentShared) randomVarianceUint(init uint, maxVariance uint) uint {
	return uint(rand.Intn(int(maxVariance)*2)) - maxVariance + init
}

func (h *SuperTestComponentShared) randomGzipLevel() uint {
	return uint(rand.Intn(9))
}

func (h *SuperTestComponentShared) randomPlacement() string {
	if h.Service.Type == ServiceTypeVCL {
		return h.randomChoiceString([]string{"none", "waf_debug"})
	}
	return "none"
}

func (h *SuperTestComponentShared) randomFormatVersion() uint {
	if h.Service.Type == ServiceTypeVCL {
		return h.randomChoiceUint([]uint{1, 2})
	}
	return 0
}

func (h *SuperTestComponentShared) randomTimestampFormat() string {
	return h.randomChoiceString([]string{
		"%Y-%m-%dT%H:%M:%S.000"})
}

func (h *SuperTestComponentShared) randomFormat() string {
	if h.Service.Type == ServiceTypeVCL {
		return h.randomChoiceString([]string{
			"%h %l %u %t %>s %b",
			"%h %l %u %t %>s %b %T"})
	}
	return ""
}

func (h *SuperTestComponentShared) randomChoiceString(a []string) string {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	return a[rand.Intn(len(a))]
}

func (h *SuperTestComponentShared) randomChoiceUint(a []uint) uint {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	return a[rand.Intn(len(a))]
}

func (h *SuperTestComponentShared) randomSelectionString(num int, s []string) []string {
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	var scopy []string
	copy(scopy, s)
	for i := len(scopy); i > num; i = i - 1 {
		randomIndex := rand.Intn(len(scopy))
		scopy = append(scopy[:randomIndex], scopy[randomIndex+1:]...)
	}
	return scopy
}
