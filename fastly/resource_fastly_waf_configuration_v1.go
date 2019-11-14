package fastly

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

var fastlyNoWAFConfigurationFoundErr = errors.New("No Matching Fastly WAF Configuration found.")

func resourceWAFConfigurationV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceWAFConfigurationV1Create,
		Read:   resourceWAFConfigurationV1Read,
		Update: resourceWAFConfigurationV1Update,
		Delete: resourceWAFConfigurationV1Delete,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Unique id of the WAF configuration resource",
			},
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "Managed by Terraform",
				Description: "A short version comment summarizing changes included in a specific firewall version.",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether a specific firewall version is currently deployed.",
			},
			"locked": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether a specific firewall version is locked from being modified.",
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
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "HTTP violation threshold.",
			},
			"inbound_anomaly_score_threshold": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Inbound anomaly threshold.",
			},
			"lfi_score_threshold": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Local file inclusion attack threshold.",
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
		},
	}
}

func resourceWAFConfigurationV1Create(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceWAFConfigurationV1Update(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceWAFConfigurationV1Read(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceWAFConfigurationV1Delete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
