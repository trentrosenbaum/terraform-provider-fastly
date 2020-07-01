package fastly

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestResourceFastlyFlattenBackend(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Backend
		local  []map[string]interface{}
	}{
		{
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
		out := flattenBackends(c.remote)
		if !reflect.DeepEqual(out, c.local) {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

type FastlyServiceV1BackendTestCase struct {
	Name         string
	DomainName   string
	Backends     []FastlyServiceV1BackendTestCaseBackend
	BackendsTest []FastlyServiceV1BackendTestCaseBackend
	DefaultTTL   int
}

type FastlyServiceV1BackendTestCaseBackend struct {
	Name    string
	Address string
}


func TestAccFastlyServiceV1_Backend(t *testing.T) {
	var service gofastly.ServiceDetail

	var name       = makeTestServiceName()
	var domainName = makeTestDomainName()
	var backends   = makeTestRandomBackendConfig(2)

	cases := map[string]FastlyServiceV1BackendTestCase{
		"backend": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:1],
		},
		"backend_update": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:2],
			DefaultTTL: 3400,
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigServiceV1_Backends(cases["backend"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["backend"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "1"),
				),
			},
			{
				Config: testResourceConfigServiceV1_BackendsTTL(cases["backend_update"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["backend_update"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "backend.#", "2"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "2"),
				),
			},
		},
	})
}

// TestAccFastlyServiceV1_BackendWasm provides Wasm with it's own test - no default_ttl allowed.
func TestAccFastlyServiceV1_BackendWasm(t *testing.T) {
	var service gofastly.ServiceDetail

	var name       = makeTestServiceName()
	var domainName = makeTestDomainName()
	var backends   = makeTestRandomBackendConfig(1)

	cases := map[string]FastlyServiceV1BackendTestCase{
		"backend": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:1],
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigServiceV1_Backends(cases["backend"], TestServiceTypeWasm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeWasm.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["backend"]),
					resource.TestCheckResourceAttr(TestServiceTypeWasm.ref, "active_version", "1"),
				),
			},
		},
	})
}

func TestAccFastlyServiceV1_BackendInvalid(t *testing.T) {
	var service gofastly.ServiceDetail

	var name = makeTestServiceName()
	var domainName = makeTestDomainName()
	var backends = makeTestRandomBackendConfig(3)

	// Make first (00) backend invalid for invalid test
	backends[0].Address = backends[0].Address + "."

	cases := map[string]FastlyServiceV1BackendTestCase{
		"vcl_backend_bad": FastlyServiceV1BackendTestCase{
			Name:         name,
			DomainName:   domainName,
			Backends:     backends[:1],
			BackendsTest: []FastlyServiceV1BackendTestCaseBackend{}, // Overrides checking of backends
		},
		"vcl_backend_bad_update": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[1:3],
			DefaultTTL: 3400,
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config:      testResourceConfigServiceV1_Backends(cases["vcl_backend_bad"], TestServiceTypeVCL),
				ExpectError: regexp.MustCompile("Bad Request"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_bad"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "1"),
				),
			},
			{
				Config: testResourceConfigServiceV1_BackendsTTL(cases["vcl_backend_bad_update"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_bad_update"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "1"),
				),
			},
		},
	})

}

func TestAccFastlyServiceV1_DefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail

	var name = makeTestServiceName()
	var domainName = makeTestDomainName()
	var backends = makeTestRandomBackendConfig(2)

	cases := map[string]FastlyServiceV1BackendTestCase{
		"vcl_backend_default_ttl": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:1],
		},
		"vcl_backend_default_ttl_update": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:2],
			DefaultTTL: 3400,
		},
		"vcl_backend_default_ttl_zero": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:2],
			DefaultTTL: 0,
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigServiceV1_Backends(cases["vcl_backend_default_ttl"],TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_default_ttl"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "default_ttl", "3600"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "1"),
				),
			},
			{
				Config: testResourceConfigServiceV1_BackendsTTL(cases["vcl_backend_default_ttl_update"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_default_ttl_update"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "default_ttl", "3400"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "2"),
				),
			},
			// Now update the default_ttl to 0 and encounter the issue https://github.com/hashicorp/terraform/issues/12910
			{
				Config: testResourceConfigServiceV1_BackendsTTL(cases["vcl_backend_default_ttl_zero"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_default_ttl_zero"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "default_ttl", "0"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "3"),
				),
			},
		},
	})

}

func TestAccFastlyServiceV1_CreateDefaultTTL(t *testing.T) {
	var service gofastly.ServiceDetail

	var name = makeTestServiceName()
	var domainName = makeTestDomainName()
	var backends = makeTestRandomBackendConfig(2)

	cases := map[string]FastlyServiceV1BackendTestCase{
		"vcl_backend_create_default_ttl": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[:1],
			DefaultTTL: 3400,
		},
		"vcl_backend_create_default_ttl_zero": FastlyServiceV1BackendTestCase{
			Name:       name,
			DomainName: domainName,
			Backends:   backends[1:2],
			DefaultTTL: 0,
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigServiceV1_BackendsTTL(cases["vcl_backend_create_default_ttl"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_create_default_ttl"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "default_ttl", "3400"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "1"),
				),
			},
			{
				Config: testResourceConfigServiceV1_BackendsTTL(cases["vcl_backend_create_default_ttl_zero"], TestServiceTypeVCL),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestServiceTypeVCL.ref, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_create_default_ttl_zero"]),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "default_ttl", "0"),
					resource.TestCheckResourceAttr(TestServiceTypeVCL.ref, "active_version", "2"),
				),
			},
		},
	})
}


func testAccCheckFastlyServiceV1Attributes_Backends(service *gofastly.ServiceDetail, c FastlyServiceV1BackendTestCase) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != c.Name {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", c.Name, service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		// Use BackendsTest if provided, default to Backends
		backendsPtr := c.BackendsTest
		if backendsPtr == nil {
			backendsPtr = c.Backends
		}

		expected := len(backendList)
		if backendsPtr != nil {
			for _, b := range backendList {
				for _, e := range backendsPtr {
					if b.Address == e.Address {
						expected--
					}
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Backend count mismatch, expected: %#v, got: %#v", c.BackendsTest, backendList)
		}

		return nil
	}
}

func makeTestRandomBackendConfig(num int) []FastlyServiceV1BackendTestCaseBackend {
	var backends = []FastlyServiceV1BackendTestCaseBackend{}
	for i := 0; i < num; i++ {
		backends = append(backends, FastlyServiceV1BackendTestCaseBackend{
			Address: fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3)),
			Name:    fmt.Sprintf("tf-test-backend-%02d", i),
		})
	}
	return backends
}

func testResourceConfigServiceV1_Backends(data FastlyServiceV1BackendTestCase, serviceType testServiceType) string {
	return testGetResourceTemplate(fmt.Sprintf("service_%s_backends", serviceType.code), data)
}

func testResourceConfigServiceV1_BackendsTTL(data FastlyServiceV1BackendTestCase, serviceType testServiceType) string {
	return testGetResourceTemplate(fmt.Sprintf("service_%s_backends_ttl", serviceType.code), data)
}
