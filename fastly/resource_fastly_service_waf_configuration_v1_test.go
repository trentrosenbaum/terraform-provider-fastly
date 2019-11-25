package fastly

import (
	"fmt"
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"testing"
)

func TestAccFastlyServiceWAFVersionV1DetermineVersion(t *testing.T) {

	cases := []struct {
		remote []*gofastly.WAFVersion
		local  int
	}{
		{
			remote: []*gofastly.WAFVersion{
				{Number: 1},
				{Number: 2},
			},
			local: 2,
		},
		{
			remote: []*gofastly.WAFVersion{
				{Number: 3},
				{Number: 2},
				{Number: 1},
			},
			local: 3,
		},
	}

	for _, c := range cases {
		out := determineLatestVersion(c.remote)
		if c.local != out.Number {
			t.Fatalf("Error matching:\nexpected: %#v\n     got: %#v", c.local, out)
		}
	}
}

func TestAccFastlyServiceWAFVersionV1Add(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput := buildTestUpdateInput()
	wafVer := composeWAFConfiguration(wafVerInput)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1AddExistingService(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput := buildTestUpdateInput()
	wafVer := composeWAFConfiguration(wafVerInput)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 1),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Update(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput1 := buildTestUpdateInput()
	wafVer1 := composeWAFConfiguration(wafVerInput1)

	wafVerInput2 := buildTestUpdateInput()
	wafVerInput2.XSSScoreThreshold = 1
	wafVer2 := composeWAFConfiguration(wafVerInput2)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput1, 1),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, wafVer2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput2, 2),
				),
			},
		},
	})
}

func TestAccFastlyServiceWAFVersionV1Delete(t *testing.T) {
	var service gofastly.ServiceDetail
	name := fmt.Sprintf("tf-test-%s", acctest.RandString(10))
	wafVerInput := buildTestUpdateInput()
	wafVer := composeWAFConfiguration(wafVerInput)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckServiceV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceV1WAFVersion(name, wafVer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 1),
				),
			},
			{
				Config: testAccServiceV1WAFVersion(name, ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServiceV1Exists(serviceRef, &service),
					testAccCheckFastlyServiceWAFVersionV1CheckAttributes(&service, wafVerInput, 2),
				),
			},
		},
	})
}

func testAccCheckFastlyServiceWAFVersionV1CheckAttributes(service *gofastly.ServiceDetail, i *gofastly.UpdateWAFVersionInput, latestVersion int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		wafResp, err := conn.ListWAFs(&gofastly.ListWAFsInput{
			FilterService: service.ID,
			FilterVersion: service.ActiveVersion.Number,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(wafResp.Items) != 1 {
			return fmt.Errorf("[ERR] Expected waf result size (%d), got (%d)", 1, len(wafResp.Items))
		}

		waf := wafResp.Items[0]
		verResp, err := conn.ListWAFVersions(&gofastly.ListWAFVersionsInput{
			WAFID: waf.ID,
		})
		if err != nil {
			return fmt.Errorf("[ERR] Error looking up WAF version records for (%s), version (%v): %s", service.Name, service.ActiveVersion.Number, err)
		}

		if len(verResp.Items) < 1 {
			return fmt.Errorf("[ERR] Expected result size (%d), got (%d)", 1, len(verResp.Items))
		}

		o, err := getVersionNumber(verResp.Items, latestVersion)
		if err != nil {
			return err
		}

		if i.AllowedHTTPVersions != o.AllowedHTTPVersions {
			return fmt.Errorf("expected AllowedHTTPVersions %s: got %s", i.AllowedHTTPVersions, o.AllowedHTTPVersions)
		}
		if i.AllowedMethods != o.AllowedMethods {
			return fmt.Errorf("expected AllowedMethods %s: got %s", i.AllowedMethods, o.AllowedMethods)
		}
		if i.AllowedRequestContentType != o.AllowedRequestContentType {
			return fmt.Errorf("expected AllowedRequestContentType %s: got %s", i.AllowedRequestContentType, o.AllowedRequestContentType)
		}
		if i.AllowedRequestContentTypeCharset != o.AllowedRequestContentTypeCharset {
			return fmt.Errorf("expected AllowedRequestContentTypeCharset %s: got %s", i.AllowedRequestContentTypeCharset, o.AllowedRequestContentTypeCharset)
		}
		if i.ArgLength != o.ArgLength {
			return fmt.Errorf("expected ArgLength %d: got %d", i.ArgLength, o.ArgLength)
		}
		if i.ArgNameLength != o.ArgNameLength {
			return fmt.Errorf("expected ArgNameLength %d: got %d", i.ArgNameLength, o.ArgNameLength)
		}
		if i.CombinedFileSizes != o.CombinedFileSizes {
			return fmt.Errorf("expected CombinedFileSizes %d: got %d", i.CombinedFileSizes, o.CombinedFileSizes)
		}
		if i.CriticalAnomalyScore != o.CriticalAnomalyScore {
			return fmt.Errorf("expected CriticalAnomalyScore %d: got %d", i.CriticalAnomalyScore, o.CriticalAnomalyScore)
		}
		if i.CRSValidateUTF8Encoding != o.CRSValidateUTF8Encoding {
			return fmt.Errorf("expected CRSValidateUTF8Encoding %v: got %v", i.CRSValidateUTF8Encoding, o.CRSValidateUTF8Encoding)
		}
		if i.ErrorAnomalyScore != o.ErrorAnomalyScore {
			return fmt.Errorf("expected ErrorAnomalyScore %d: got %d", i.ErrorAnomalyScore, o.ErrorAnomalyScore)
		}
		if i.HighRiskCountryCodes != o.HighRiskCountryCodes {
			return fmt.Errorf("expected HighRiskCountryCodes %s: got %s", i.HighRiskCountryCodes, o.HighRiskCountryCodes)
		}
		if i.HTTPViolationScoreThreshold != o.HTTPViolationScoreThreshold {
			return fmt.Errorf("expected HTTPViolationScoreThreshold %d: got %d", i.HTTPViolationScoreThreshold, o.HTTPViolationScoreThreshold)
		}
		if i.InboundAnomalyScoreThreshold != o.InboundAnomalyScoreThreshold {
			return fmt.Errorf("expected InboundAnomalyScoreThreshold %d: got %d", i.InboundAnomalyScoreThreshold, o.InboundAnomalyScoreThreshold)
		}
		if i.LFIScoreThreshold != o.LFIScoreThreshold {
			return fmt.Errorf("expected LFIScoreThreshold %d: got %d", i.LFIScoreThreshold, o.LFIScoreThreshold)
		}
		if i.MaxFileSize != o.MaxFileSize {
			return fmt.Errorf("expected MaxFileSize %d: got %d", i.MaxFileSize, o.MaxFileSize)
		}
		if i.MaxNumArgs != o.MaxNumArgs {
			return fmt.Errorf("expected MaxNumArgs %d: got %d", i.MaxNumArgs, o.MaxNumArgs)
		}
		if i.NoticeAnomalyScore != o.NoticeAnomalyScore {
			return fmt.Errorf("expected NoticeAnomalyScore %d: got %d", i.NoticeAnomalyScore, o.NoticeAnomalyScore)
		}
		if i.ParanoiaLevel != o.ParanoiaLevel {
			return fmt.Errorf("expected ParanoiaLevel %d: got %d", i.ParanoiaLevel, o.ParanoiaLevel)
		}
		if i.PHPInjectionScoreThreshold != o.PHPInjectionScoreThreshold {
			return fmt.Errorf("expected PHPInjectionScoreThreshold %d: got %d", i.PHPInjectionScoreThreshold, o.PHPInjectionScoreThreshold)
		}
		if i.RCEScoreThreshold != o.RCEScoreThreshold {
			return fmt.Errorf("expected RCEScoreThreshold %d: got %d", i.RCEScoreThreshold, o.RCEScoreThreshold)
		}
		if i.RestrictedExtensions != o.RestrictedExtensions {
			return fmt.Errorf("expected RestrictedExtensions %s: got %s", i.RestrictedExtensions, o.RestrictedExtensions)
		}
		if i.RestrictedHeaders != o.RestrictedHeaders {
			return fmt.Errorf("expected RestrictedHeaders %s: got %s", i.RestrictedHeaders, o.RestrictedHeaders)
		}
		if i.RFIScoreThreshold != o.RFIScoreThreshold {
			return fmt.Errorf("expected RFIScoreThreshold %d: got %d", i.RFIScoreThreshold, o.RFIScoreThreshold)
		}
		if i.SessionFixationScoreThreshold != o.SessionFixationScoreThreshold {
			return fmt.Errorf("expected SessionFixationScoreThreshold %d: got %d", i.SessionFixationScoreThreshold, o.SessionFixationScoreThreshold)
		}
		if i.SQLInjectionScoreThreshold != o.SQLInjectionScoreThreshold {
			return fmt.Errorf("expected SQLInjectionScoreThreshold %d: got %d", i.SQLInjectionScoreThreshold, o.SQLInjectionScoreThreshold)
		}
		if i.TotalArgLength != o.TotalArgLength {
			return fmt.Errorf("expected TotalArgLength %d: got %d", i.TotalArgLength, o.TotalArgLength)
		}
		if i.WarningAnomalyScore != o.WarningAnomalyScore {
			return fmt.Errorf("expected WarningAnomalyScore %d: got %d", i.WarningAnomalyScore, o.WarningAnomalyScore)
		}
		if i.XSSScoreThreshold != o.XSSScoreThreshold {
			return fmt.Errorf("expected XSSScoreThreshold %d: got %d", i.XSSScoreThreshold, o.XSSScoreThreshold)
		}
		return nil
	}
}

func getVersionNumber(versions []*gofastly.WAFVersion, number int) (gofastly.WAFVersion, error) {
	for _, v := range versions {
		if v.Number == number {
			return *v, nil
		}
	}
	return gofastly.WAFVersion{}, fmt.Errorf("version number %d not found", number)
}

func composeWAFConfiguration(i *gofastly.UpdateWAFVersionInput) string {

	return fmt.Sprintf(`
		resource "fastly_service_waf_configuration_v1" "waf" {
  			waf_id = fastly_service_v1.foo.waf[0].waf_id
            comment = "%s"
  			crs_validate_utf8_encoding = %v
  			allowed_http_versions = "%s"
            allowed_methods = "%s"
            allowed_request_content_type = "%s"
            allowed_request_content_type_charset = "%s"
            high_risk_country_codes = "%s"
            restricted_extensions = "%s"
            restricted_headers = "%s"
            arg_length = %d
            arg_name_length = %d
            combined_file_sizes = %d
            critical_anomaly_score = %d
            error_anomaly_score = %d
            http_violation_score_threshold = %d
            inbound_anomaly_score_threshold = %d
            lfi_score_threshold = %d
            max_file_size = %d
            max_num_args = %d
            notice_anomaly_score = %d
            paranoia_level = %d
            php_injection_score_threshold = %d
            rce_score_threshold = %d
            rfi_score_threshold = %d
            session_fixation_score_threshold = %d
            sql_injection_score_threshold = %d
            total_arg_length = %d
            warning_anomaly_score = %d
            xss_score_threshold = %d
}`, i.Comment, i.CRSValidateUTF8Encoding, i.AllowedHTTPVersions, i.AllowedMethods, i.AllowedRequestContentType,
		i.AllowedRequestContentTypeCharset, i.HighRiskCountryCodes, i.RestrictedExtensions, i.RestrictedHeaders,
		i.ArgLength, i.ArgNameLength, i.CombinedFileSizes, i.CriticalAnomalyScore, i.ErrorAnomalyScore, i.HTTPViolationScoreThreshold,
		i.InboundAnomalyScoreThreshold, i.LFIScoreThreshold, i.MaxFileSize, i.MaxNumArgs, i.NoticeAnomalyScore, i.ParanoiaLevel,
		i.PHPInjectionScoreThreshold, i.RCEScoreThreshold, i.RFIScoreThreshold, i.SessionFixationScoreThreshold,
		i.SQLInjectionScoreThreshold, i.TotalArgLength, i.WarningAnomalyScore, i.XSSScoreThreshold)
}

func testAccServiceV1WAFVersion(name, extraHCL string) string {

	backendName := fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3))
	domainName := fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))

	return fmt.Sprintf(`
resource "fastly_service_v1" "foo" {
  name = "%s"

  domain {
    name    = "%s"
    comment = "tf-testing-domain"
  }

  backend {
    address = "%s"
    name    = "tf -test backend"
  }

  condition {
	name = "prefetch"
	type = "PREFETCH"
	statement = "req.url~+\"index.html\""
  }

  response_object {
	name = "response"
	status = "403"
	response = "Forbidden"
	content = "content"
  }

  waf { 
	prefetch_condition = "prefetch" 
	response_object = "response"
  }

  force_destroy = true
}
  %s
`, name, domainName, backendName, extraHCL)
}

func buildTestUpdateInput() *gofastly.UpdateWAFVersionInput {
	return &gofastly.UpdateWAFVersionInput{
		Comment:                          "my comment",
		AllowedHTTPVersions:              "HTTP/1.0 HTTP/1.1",
		AllowedMethods:                   "GET HEAD POST",
		AllowedRequestContentType:        "application/x-www-form-urlencoded|multipart/form-data|text/xml|application/xml",
		AllowedRequestContentTypeCharset: "utf-8|iso-8859-1",
		ArgLength:                        800,
		ArgNameLength:                    200,
		CombinedFileSizes:                20000000,
		CriticalAnomalyScore:             12,
		CRSValidateUTF8Encoding:          true,
		ErrorAnomalyScore:                10,
		HighRiskCountryCodes:             "gb",
		HTTPViolationScoreThreshold:      20,
		InboundAnomalyScoreThreshold:     20,
		LFIScoreThreshold:                20,
		MaxFileSize:                      20000000,
		MaxNumArgs:                       510,
		NoticeAnomalyScore:               8,
		ParanoiaLevel:                    2,
		PHPInjectionScoreThreshold:       20,
		RCEScoreThreshold:                20,
		RestrictedExtensions:             ".asa/ .asax/ .ascx/ .axd/ .backup/ .bak/ .bat/ .cdx/ .cer/ .cfg/ .cmd/ .com/",
		RestrictedHeaders:                "/proxy/ /lock-token/",
		RFIScoreThreshold:                20,
		SessionFixationScoreThreshold:    20,
		SQLInjectionScoreThreshold:       20,
		TotalArgLength:                   12800,
		WarningAnomalyScore:              20,
		XSSScoreThreshold:                20,
	}
}
