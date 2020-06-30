package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"reflect"
	"regexp"
	"testing"
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


func TestAccFastlyServiceV1_Backend(t *testing.T) {
	var service gofastly.ServiceDetail

	var setups = map[string] map[string] string{}
	for _, s := range []string{"backend", "backend_bad", "default_ttl", "create_default_ttl", "create_default_ttl_zero"} {
		setups[s] = map[string] string {
			"name": 	makeTestServiceName(),
			"domain":   makeTestDomainName(),
		}
	}

	var backends = []map[string]interface{}{}
	for i := 0; i < 9; i++ {
		backends = append(backends, map[string]interface{}{
			"address": fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3)),
			"name":    fmt.Sprintf("tf-test-backend-%02d", i),
		})
	}

	// Make third (02) backend invalid for invalid test
	backends[2]["address"] = backends[2]["address"].(string) + "."

	cases := map[string] map[string] interface{}{
		"vcl_backend": {
			"name":        		setups["backend"]["name"],
			"domain_name": 		setups["backend"]["domain"],
			"backends": 		[]map[string]interface{}{backends[0]},
		},
		"vcl_backend_update": {
			"name":        		setups["backend"]["name"],
			"domain_name": 		setups["backend"]["domain"],
			"backends":			[]map[string]interface{}{backends[0], backends[1]},
			"default_ttl":		3400,
		},
		"vcl_backend_bad": {
			"name":        		setups["backend_bad"]["name"],
			"domain_name": 		setups["backend_bad"]["domain"],
			"backends":			[]map[string]interface{}{backends[2]},
			"backends_check":   []map[string]interface{}{}, // Overrides checking of backends
		},
		"vcl_backend_bad_update": {
			"name":        		setups["backend_bad"]["name"],
			"domain_name": 		setups["backend_bad"]["domain"],
			"backends":			[]map[string]interface{}{backends[3],backends[4]},
			"default_ttl":		3400,
		},
		"vcl_backend_default_ttl": {
			"name":        		setups["default_ttl"]["name"],
			"domain_name": 		setups["default_ttl"]["domain"],
			"backends":			[]map[string]interface{}{backends[5]},
		},
		"vcl_backend_default_ttl_update": {
			"name":        		setups["default_ttl"]["name"],
			"domain_name": 		setups["default_ttl"]["domain"],
			"backends":			[]map[string]interface{}{backends[5],backends[6]},
			"default_ttl":		3400,
		},
		"vcl_backend_default_ttl_zero": {
			"name":        		setups["default_ttl"]["name"],
			"domain_name": 		setups["default_ttl"]["domain"],
			"backends":			[]map[string]interface{}{backends[5],backends[6]},
			"default_ttl":		0,
		},
		"vcl_backend_create_default_ttl": {
			"name":        		setups["create_default_ttl"]["name"],
			"domain_name": 		setups["create_default_ttl"]["domain"],
			"backends":			[]map[string]interface{}{backends[7]},
			"default_ttl":		3400,
		},
		"vcl_backend_create_default_ttl_zero": {
			"name":        		setups["create_default_ttl_zero"]["name"],
			"domain_name": 		setups["create_default_ttl_zero"]["domain"],
			"backends":			[]map[string]interface{}{backends[8]},
			"default_ttl":		0,
		},

	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend"]),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_update"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_update"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef,"active_version", "2"),
					resource.TestCheckResourceAttr(TestVCLServiceRef,"backend.#", "2"),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_bad"]),
				ExpectError: regexp.MustCompile("Bad Request"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_bad"]),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_bad_update"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_bad_update"]),
				),
			},

			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_default_ttl"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_default_ttl"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "default_ttl", "3600"),
				),
			},

			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_default_ttl_update"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_default_ttl"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "default_ttl", "3400"),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "active_version", "2"),
				),
			},
			// Now update the default_ttl to 0 and encounter the issue https://github.com/hashicorp/terraform/issues/12910
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_default_ttl_zero"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_default_ttl_zero"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "default_ttl", "0"),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "active_version", "3"),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_create_default_ttl"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_create_default_ttl"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "default_ttl", "3400"),
				),
			},
			{
				Config: testResourceConfigVCLServiceV1_Backends(cases["vcl_backend_create_default_ttl_zero"]),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(TestVCLServiceRef, &service),
					testAccCheckFastlyServiceV1Attributes_Backends(&service, cases["vcl_backend_create_default_ttl_zero"]),
					resource.TestCheckResourceAttr(TestVCLServiceRef, "default_ttl", "0"),
				),
			},
		},
	})

}





func testAccCheckFastlyServiceV1Attributes_Backends(service *gofastly.ServiceDetail, c map[string] interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if service.Name != c["name"] {
			return fmt.Errorf("Bad name, expected (%s), got (%s)", c["name"], service.Name)
		}

		conn := testAccProvider.Meta().(*FastlyClient).conn
		backendList, err := conn.ListBackends(&gofastly.ListBackendsInput{
			Service: service.ID,
			Version: service.ActiveVersion.Number,
		})

		if err != nil {
			return fmt.Errorf("[ERR] Error looking up Backends for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		var backendPtr interface{}
		var checkBackends []map[string] interface{}

		if v, ok := c["backends_check"]; ok {
			backendPtr = v
		} else {
			backendPtr = c["backends"]
		}

		switch t := backendPtr.(type) {
		case []map[string]interface{}:
			checkBackends = t
		}

		expected := len(backendList)
		if checkBackends!=nil {
			for _, b := range backendList {
				for _, e := range checkBackends {
					if b.Address == e["address"] {
						expected--
					}
				}
			}
		}

		if expected > 0 {
			return fmt.Errorf("Backend count mismatch, expected: %#v, got: %#v", checkBackends, backendList)
		}

		return nil
	}
}


func testResourceConfigVCLServiceV1_Backends(data map[string] interface{}) string {
	return testGetResourceTemplate("service_vcl_backends", data)
}

