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

// SuperTestBackendData provides a data structure for multiple Backend blocks.
type SuperTestBackendData struct {
	Backends []*gofastly.Backend
}

// SuperTestComponentBackend mixes-in the shared SuperTestComponent functions with SuperTestBackendData.
type SuperTestComponentBackend struct {
	*SuperTestComponentShared
	*SuperTestBackendData
}

// NewSuperTestComponentBackend is the constructor for a SuperTestComponentBackend.
// It injects the service metadata and creates a single random SuperTestBackendBlock.
func NewSuperTestComponentBackend(service *SuperTestService) *SuperTestComponentBackend {
	var c SuperTestComponentBackend
	c.SuperTestComponentShared = &SuperTestComponentShared{
		service,
	}
	c.SuperTestBackendData = &SuperTestBackendData{
		[]*gofastly.Backend{c.randomBackend()},
	}
	return &c
}

// randomBackend creates a random SuperTestBackendBlock.
func (c *SuperTestComponentBackend) randomBackend() *gofastly.Backend {
	return &gofastly.Backend{
		Name:                fmt.Sprintf("tf-test-backend-%s", acctest.RandString(5)),
		Address:             fmt.Sprintf("%s.tf-test.fastly.com", acctest.RandString(5)),
		AutoLoadbalance:     true,
		BetweenBytesTimeout: 10000,
		ConnectTimeout:      1000,
		ErrorThreshold:      0,
		FirstByteTimeout:    15000,
		MaxConn:             200,
		Port:                80,
		OverrideHost:        "",
		RequestCondition:    "",
		UseSSL:              false,
		MaxTLSVersion:       "",
		MinTLSVersion:       "",
		SSLCiphers:          []string{},
		SSLCACert:           "",
		SSLClientCert:       "",
		SSLClientKey:        "",
		SSLCheckCert:        true,
		SSLHostname:         "",
		SSLCertHostname:     "",
		SSLSNIHostname:      "",
		Shield:              "",
		Weight:              101,
		HealthCheck:         "",
	}
}

// AddRandom fulfils the SuperTestComponent interface.
// It appends a random SuperTestBackendBlock to the list.
func (c *SuperTestComponentBackend) AddRandom() error {
	c.Backends = append(c.Backends, c.randomBackend())
	return nil
}

// UpdateRandom fulfils the SuperTestComponent interface.
// It updates a randomly selected SuperTestBackendBlock to a new random set of values.
func (c *SuperTestComponentBackend) UpdateRandom() error {
	var randomIndex = c.randomIndex()
	c.Backends[randomIndex] = c.randomBackend()
	return nil
}

// DeleteRandom fulfils the SuperTestComponent interface.
// It deletes a random SuperTestBackendBlock.
func (c *SuperTestComponentBackend) DeleteRandom() error {
	var numBackends = len(c.Backends)
	if numBackends < 2 {
		return fmt.Errorf("cannot delete random backend only have %v", numBackends)
	}
	var randomIndex = c.randomIndex()
	c.Backends = append(c.Backends[:randomIndex], c.Backends[randomIndex+1:]...)
	return nil
}

// randomIndex selects an index at random from the list of SuperTestBackendBlocks.
func (c *SuperTestComponentBackend) randomIndex() int {
	return rand.Intn(len(c.Backends))
}

// GetTestCheckFunc fulfils the SuperTestComponent interface.
// It composes a series of tests to verify this component of a provisioned resource against
// the data used to provision it.
func (c SuperTestComponentBackend) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc {
	var numBackends = len(c.Backends)
	var r = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "backend.#", strconv.Itoa(numBackends)),
		c.testState(),
		c.testApi(serviceDetail),
	}
	return r
}

// testState checks a provisioned resource Backend *state* against the data used to provision it.
func (c SuperTestComponentBackend) testState() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var serviceMetadata = c.getServiceMetadata()
		var sourceList = flattenBackend(c.Backends, *serviceMetadata)
		var targetList, err = c.getStateTypeSetBlocks(s, "backend")
		if err != nil {
			return err
		}

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("state backend mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testApi checks a provisioned resource Backend at the *api* against the data used to provision it.
func (c SuperTestComponentBackend) testApi(serviceDetail *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		apiList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			Service: serviceDetail.ID,
			Version: serviceDetail.Version.Number,
		})
		if err != nil {
			return err
		}

		var serviceMetadata = c.getServiceMetadata()
		var sourceList = flattenBackend(c.Backends, *serviceMetadata)
		var targetList = flattenBackend(apiList, *serviceMetadata)

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("api backend mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testEquivalent compares two flattened data Backends
func (c *SuperTestComponentBackend) testEquivalent(ms1 []map[string]interface{}, ms2 []map[string]interface{}) bool {
	expected := len(ms1)
	for _, m1 := range ms1 {
		for _, m2 := range ms2 {
			if c.compareKey(m1, m2, "name") &&
				c.compareKey(m1, m2, "address") &&
				c.compareKey(m1, m2, "auto_loadbalance") &&
				c.compareKey(m1, m2, "between_bytes_timeout") &&
				c.compareKey(m1, m2, "connect_timeout") &&
				c.compareKey(m1, m2, "error_threshold") &&
				c.compareKey(m1, m2, "first_byte_timeout") &&
				c.compareKey(m1, m2, "max_conn") &&
				c.compareKey(m1, m2, "port") &&
				c.compareKey(m1, m2, "override_host") &&
				c.compareKey(m1, m2, "request_condition") &&
				c.compareKey(m1, m2, "use_ssl") &&
				c.compareKey(m1, m2, "max_tls_version") &&
				c.compareKey(m1, m2, "min_tls_version") &&
				c.compareKey(m1, m2, "ssl_ciphers") &&
				c.compareKey(m1, m2, "ssl_ca_cert") &&
				c.compareKey(m1, m2, "ssl_client_cert") &&
				c.compareKey(m1, m2, "ssl_client_key") &&
				c.compareKey(m1, m2, "ssl_check_cert") &&
				c.compareKey(m1, m2, "ssl_hostname") &&
				c.compareKey(m1, m2, "ssl_cert_hostname") &&
				c.compareKey(m1, m2, "ssl_sni_hostname") &&
				c.compareKey(m1, m2, "shield") &&
				c.compareKey(m1, m2, "weight") &&
				c.compareKey(m1, m2, "healthcheck") {
				expected--
			}
		}
	}
	return expected <= 0
}

func TestResourceFastlyFlattenBackend(t *testing.T) {
	cases := []struct {
		serviceMetadata ServiceMetadata
		remote          []*gofastly.Backend
		local           []map[string]interface{}
	}{
		{
			serviceMetadata: ServiceMetadata{
				serviceType: ServiceTypeVCL,
			},
			remote: []*gofastly.Backend{
				{
					Name:                "test.notexample.com",
					Address:             "www.notexample.com",
					OverrideHost:        "origin.example.com",
					Port:                uint(80),
					AutoLoadbalance:     true,
					BetweenBytesTimeout: uint(10000),
					ConnectTimeout:      uint(1000),
					ErrorThreshold:      uint(0),
					FirstByteTimeout:    uint(15000),
					MaxConn:             uint(200),
					RequestCondition:    "",
					HealthCheck:         "",
					UseSSL:              false,
					SSLCheckCert:        true,
					SSLHostname:         "",
					SSLCACert:           "",
					SSLCertHostname:     "",
					SSLSNIHostname:      "",
					SSLClientKey:        "",
					SSLClientCert:       "",
					MaxTLSVersion:       "",
					MinTLSVersion:       "",
					SSLCiphers:          []string{"foo", "bar", "baz"},
					Shield:              "New York",
					Weight:              uint(100),
				},
			},
			local: []map[string]interface{}{
				{
					"name":                  "test.notexample.com",
					"address":               "www.notexample.com",
					"override_host":         "origin.example.com",
					"port":                  80,
					"auto_loadbalance":      true,
					"between_bytes_timeout": 10000,
					"connect_timeout":       1000,
					"error_threshold":       0,
					"first_byte_timeout":    15000,
					"max_conn":              200,
					"request_condition":     "",
					"healthcheck":           "",
					"use_ssl":               false,
					"ssl_check_cert":        true,
					"ssl_hostname":          "",
					"ssl_ca_cert":           "",
					"ssl_cert_hostname":     "",
					"ssl_sni_hostname":      "",
					"ssl_client_key":        "",
					"ssl_client_cert":       "",
					"max_tls_version":       "",
					"min_tls_version":       "",
					"ssl_ciphers":           "foo,bar,baz",
					"shield":                "New York",
					"weight":                100,
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenBackend(c.remote, c.serviceMetadata)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}
