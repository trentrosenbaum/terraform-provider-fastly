package fastly

import (
	"errors"
	"fmt"
	"log"
	"sort"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceServiceWAFConfigurationV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceWAFConfigurationV1Create,
		Read:   resourceServiceWAFConfigurationV1Read,
		Update: resourceServiceWAFConfigurationV1Update,
		Delete: resourceServiceWAFConfigurationV1Delete,

		Schema: map[string]*schema.Schema{
			"waf_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The service the WAF belongs to.",
			},
			"allowed_http_versions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed HTTP versions (default HTTP/1.0 HTTP/1.1 HTTP/2).",
			},
			"allowed_methods": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of HTTP method names (default GET HEAD POST OPTIONS PUT PATCH DELETE).",
			},
			"allowed_request_content_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed request content types (default application/x-www-form-urlencoded|multipart/form-data|text/xml|application/xml|application/x-amf|application/json|text/plain).",
			},
			"allowed_request_content_type_charset": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Allowed request content type charset (default utf-8|iso-8859-1|iso-8859-15|windows-1252).",
			},
			"arg_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of arguments allowed (default 400).",
			},
			"arg_name_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum allowed argument name length (default 100).",
			},
			"combined_file_sizes": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum allowed size of all files (in bytes, default 10000000).",
			},
			"critical_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for critical anomalies (default 6).",
			},
			"crs_validate_utf8_encoding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "CRS validate UTF8 encoding.",
			},
			"error_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for error anomalies (default 5).",
			},
			"high_risk_country_codes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of country codes in ISO 3166-1 (two-letter) format.",
			},
			"http_violation_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "HTTP violation threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"inbound_anomaly_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Inbound anomaly threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"lfi_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Local file inclusion attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"max_file_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum allowed file size, in bytes (default 10000000).",
			},
			"max_num_args": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum number of arguments allowed (default 255).",
			},
			"notice_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for notice anomalies (default 4).",
			},
			"paranoia_level": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The configured paranoia level (default 1).",
			},
			"php_injection_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "PHP injection threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"rce_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Remote code execution threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"restricted_extensions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of allowed file extensions (default .asa/ .asax/ .ascx/ .axd/ .backup/ .bak/ .bat/ .cdx/ .cer/ .cfg/ .cmd/ .com/ .config/ .conf/ .cs/ .csproj/ .csr/ .dat/ .db/ .dbf/ .dll/ .dos/ .htr/ .htw/ .ida/ .idc/ .idq/ .inc/ .ini/ .key/ .licx/ .lnk/ .log/ .mdb/ .old/ .pass/ .pdb/ .pol/ .printer/ .pwd/ .resources/ .resx/ .sql/ .sys/ .vb/ .vbs/ .vbproj/ .vsdisco/ .webinfo/ .xsd/ .xsx).",
			},
			"restricted_headers": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A space-separated list of allowed header names (default /proxy/ /lock-token/ /content-range/ /translate/ /if/).",
			},
			"rfi_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Remote file inclusion attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"session_fixation_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "Session fixation attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"sql_injection_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "SQL injection attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"total_arg_length": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The maximum size of argument names and values (default 6400).",
			},
			"warning_anomaly_score": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Score value to add for warning anomalies.",
			},
			"xss_score_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Description:  "XSS attack threshold.",
				ValidateFunc: validation.IntAtLeast(1),
			},
			"rule": activeRule,
		},
	}
}

// this method calls update because the creation of the waf (within the service resource) automatically creates
// the first waf version, and this makes both a create and an updating exactly the same operation.
func resourceServiceWAFConfigurationV1Create(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] creating configuration for WAF: %s", d.Get("waf_id").(string))
	return resourceServiceWAFConfigurationV1Update(d, meta)
}

func resourceServiceWAFConfigurationV1Update(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	latestVersion, err := getLatestVersion(d, meta)
	if err != nil {
		return err
	}

	wafID := d.Get("waf_id").(string)
	log.Printf("[INFO] updating configuration for WAF: %s", wafID)
	if latestVersion.Locked {
		latestVersion, err = conn.CloneWAFVersion(&gofastly.CloneWAFVersionInput{
			WAFID:            wafID,
			WAFVersionNumber: latestVersion.Number,
		})
		if err != nil {
			return err
		}
	}

	input := buildUpdateInput(d, latestVersion.ID, latestVersion.Number)
	latestVersion, err = conn.UpdateWAFVersion(input)
	if err != nil {
		return err
	}

	if d.HasChange("rule") {
		if err := updateRules(d, meta, wafID, latestVersion.Number); err != nil {
			return err
		}
	}

	err = conn.DeployWAFVersion(&gofastly.DeployWAFVersionInput{
		WAFID:            wafID,
		WAFVersionNumber: latestVersion.Number,
	})
	if err != nil {
		return err
	}

	return resourceServiceWAFConfigurationV1Read(d, meta)
}

func resourceServiceWAFConfigurationV1Read(d *schema.ResourceData, meta interface{}) error {

	latestVersion, err := getLatestVersion(d, meta)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode == 404 {
			log.Printf("[DEBUG] WAF (%s) was not found - removing from state", d.Get("waf_id").(string))
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[INFO] retrieving WAF version number: %d", latestVersion.Number)
	refreshWAFConfig(d, latestVersion)

	if err := readWAFRules(meta, d, latestVersion.Number); err != nil {
		return err
	}

	return nil
}

func resourceServiceWAFConfigurationV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	wafID := d.Get("waf_id").(string)
	log.Printf("[INFO] destroying configuration by creating empty version of WAF: %s", wafID)
	emptyVersion, err := conn.CreateEmptyWAFVersion(&gofastly.CreateEmptyWAFVersionInput{
		WAFID: wafID,
	})
	if err != nil {
		return err
	}

	err = conn.DeployWAFVersion(&gofastly.DeployWAFVersionInput{
		WAFID:            wafID,
		WAFVersionNumber: emptyVersion.Number,
	})
	if err != nil {
		return err
	}
	d.SetId("")
	return nil
}

func getLatestVersion(d *schema.ResourceData, meta interface{}) (*gofastly.WAFVersion, error) {
	conn := meta.(*FastlyClient).conn

	wafID := d.Get("waf_id").(string)
	resp, err := conn.ListAllWAFVersions(&gofastly.ListAllWAFVersionsInput{
		WAFID: wafID,
	})
	if err != nil {
		return nil, err
	}

	latest, err := determineLatestVersion(resp.Items)
	if err != nil {
		return nil, fmt.Errorf("[ERR] Error looking up WAF id: %s, with error %s", wafID, err)
	}
	return latest, nil
}

func buildUpdateInput(d *schema.ResourceData, id string, number int) *gofastly.UpdateWAFVersionInput {
	input := &gofastly.UpdateWAFVersionInput{
		WAFVersionID:     id,
		WAFVersionNumber: number,
		WAFID:            d.Get("waf_id").(string),
	}
	if v, ok := d.GetOk("allowed_http_versions"); ok {
		input.AllowedHTTPVersions = v.(string)
	}
	if v, ok := d.GetOk("allowed_methods"); ok {
		input.AllowedMethods = v.(string)
	}
	if v, ok := d.GetOk("allowed_methods"); ok {
		input.AllowedMethods = v.(string)
	}
	if v, ok := d.GetOk("allowed_request_content_type"); ok {
		input.AllowedRequestContentType = v.(string)
	}
	if v, ok := d.GetOk("allowed_request_content_type_charset"); ok {
		input.AllowedRequestContentTypeCharset = v.(string)
	}
	if v, ok := d.GetOk("arg_length"); ok {
		input.ArgLength = v.(int)
	}
	if v, ok := d.GetOk("arg_name_length"); ok {
		input.ArgNameLength = v.(int)
	}
	if v, ok := d.GetOk("combined_file_sizes"); ok {
		input.CombinedFileSizes = v.(int)
	}
	if v, ok := d.GetOk("critical_anomaly_score"); ok {
		input.CriticalAnomalyScore = v.(int)
	}
	if v, ok := d.GetOkExists("crs_validate_utf8_encoding"); ok {
		input.CRSValidateUTF8Encoding = v.(bool)
	}
	if v, ok := d.GetOk("error_anomaly_score"); ok {
		input.ErrorAnomalyScore = v.(int)
	}
	if v, ok := d.GetOk("high_risk_country_codes"); ok {
		input.HighRiskCountryCodes = v.(string)
	}
	if v, ok := d.GetOk("http_violation_score_threshold"); ok {
		input.HTTPViolationScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("inbound_anomaly_score_threshold"); ok {
		input.InboundAnomalyScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("lfi_score_threshold"); ok {
		input.LFIScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("max_file_size"); ok {
		input.MaxFileSize = v.(int)
	}
	if v, ok := d.GetOk("max_num_args"); ok {
		input.MaxNumArgs = v.(int)
	}
	if v, ok := d.GetOk("notice_anomaly_score"); ok {
		input.NoticeAnomalyScore = v.(int)
	}
	if v, ok := d.GetOk("paranoia_level"); ok {
		input.ParanoiaLevel = v.(int)
	}
	if v, ok := d.GetOk("php_injection_score_threshold"); ok {
		input.PHPInjectionScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("rce_score_threshold"); ok {
		input.RCEScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("restricted_extensions"); ok {
		input.RestrictedExtensions = v.(string)
	}
	if v, ok := d.GetOk("restricted_headers"); ok {
		input.RestrictedHeaders = v.(string)
	}
	if v, ok := d.GetOk("rfi_score_threshold"); ok {
		input.RFIScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("session_fixation_score_threshold"); ok {
		input.SessionFixationScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("sql_injection_score_threshold"); ok {
		input.SQLInjectionScoreThreshold = v.(int)
	}
	if v, ok := d.GetOk("total_arg_length"); ok {
		input.TotalArgLength = v.(int)
	}
	if v, ok := d.GetOk("warning_anomaly_score"); ok {
		input.WarningAnomalyScore = v.(int)
	}
	if v, ok := d.GetOk("xss_score_threshold"); ok {
		input.XSSScoreThreshold = v.(int)
	}
	return input
}

func refreshWAFConfig(d *schema.ResourceData, version *gofastly.WAFVersion) {

	d.SetId(version.ID)
	if v, ok := d.GetOk("allowed_http_versions"); ok {
		d.Set("allowed_http_versions", v)
	}
	if v, ok := d.GetOk("allowed_methods"); ok {
		d.Set("allowed_methods", v)
	}
	if v, ok := d.GetOk("allowed_methods"); ok {
		d.Set("allowed_methods", v)
	}
	if v, ok := d.GetOk("allowed_request_content_type"); ok {
		d.Set("allowed_request_content_type", v)
	}
	if v, ok := d.GetOk("allowed_request_content_type_charset"); ok {
		d.Set("allowed_request_content_type_charset", v)
	}
	if v, ok := d.GetOk("arg_length"); ok {
		d.Set("arg_length", v)
	}
	if v, ok := d.GetOk("arg_name_length"); ok {
		d.Set("arg_name_length", v)
	}
	if v, ok := d.GetOk("combined_file_sizes"); ok {
		d.Set("combined_file_sizes", v)
	}
	if v, ok := d.GetOk("critical_anomaly_score"); ok {
		d.Set("critical_anomaly_score", v)
	}
	if v, ok := d.GetOk("crs_validate_utf8_encoding"); ok {
		d.Set("crs_validate_utf8_encoding", v)
	}
	if v, ok := d.GetOk("error_anomaly_score"); ok {
		d.Set("error_anomaly_score", v)
	}
	if v, ok := d.GetOk("high_risk_country_codes"); ok {
		d.Set("high_risk_country_codes", v)
	}
	if v, ok := d.GetOk("http_violation_score_threshold"); ok {
		d.Set("http_violation_score_threshold", v)
	}
	if v, ok := d.GetOk("inbound_anomaly_score_threshold"); ok {
		d.Set("inbound_anomaly_score_threshold", v)
	}
	if v, ok := d.GetOk("lfi_score_threshold"); ok {
		d.Set("lfi_score_threshold", v)
	}
	if v, ok := d.GetOk("max_file_size"); ok {
		d.Set("max_file_size", v)
	}
	if v, ok := d.GetOk("max_num_args"); ok {
		d.Set("max_num_args", v)
	}
	if v, ok := d.GetOk("notice_anomaly_score"); ok {
		d.Set("notice_anomaly_score", v)
	}
	if v, ok := d.GetOk("paranoia_level"); ok {
		d.Set("paranoia_level", v)
	}
	if v, ok := d.GetOk("php_injection_score_threshold"); ok {
		d.Set("php_injection_score_threshold", v)
	}
	if v, ok := d.GetOk("rce_score_threshold"); ok {
		d.Set("rce_score_threshold", v)
	}
	if v, ok := d.GetOk("restricted_extensions"); ok {
		d.Set("restricted_extensions", v)
	}
	if v, ok := d.GetOk("restricted_headers"); ok {
		d.Set("restricted_headers", v)
	}
	if v, ok := d.GetOk("rfi_score_threshold"); ok {
		d.Set("rfi_score_threshold", v)
	}
	if v, ok := d.GetOk("session_fixation_score_threshold"); ok {
		d.Set("session_fixation_score_threshold", v)
	}
	if v, ok := d.GetOk("sql_injection_score_threshold"); ok {
		d.Set("sql_injection_score_threshold", v)
	}
	if v, ok := d.GetOk("total_arg_length"); ok {
		d.Set("total_arg_length", v)
	}
	if v, ok := d.GetOk("warning_anomaly_score"); ok {
		d.Set("warning_anomaly_score", v)
	}
	if v, ok := d.GetOk("xss_score_threshold"); ok {
		d.Set("xss_score_threshold", v)
	}
}

func determineLatestVersion(versions []*gofastly.WAFVersion) (*gofastly.WAFVersion, error) {

	if len(versions) == 0 {
		return nil, errors.New("the list of WAFVersions cannot be empty")
	}

	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Number > versions[j].Number
	})

	return versions[0], nil
}
