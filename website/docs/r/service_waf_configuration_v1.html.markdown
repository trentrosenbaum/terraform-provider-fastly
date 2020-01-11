---
layout: "fastly"
page_title: "Fastly: service_waf_configuration_v1"
sidebar_current: "docs-fastly-resource-service-waf-configuration-v1"
description: |-
  Provides a Web Application Firewall configuration and rules that can be applied to a service.
---

-> **Note:** This page is about v1.1.0 and later of the Fastly terraform provider.

# fastly_service_waf_configuration_v1

Defines Web Application Firewall configuration options that can be used to populate a service waf. This resource will populate a WAF with configuration options and WAF rules and will track their state.


~> **Warning:** Terraform will take precedence over any changes you make in the UI or API. Such changes are likely to be reversed if you run Terraform again.    


## Example Usage

Basic usage:

```hcl
variable "type_status" {
  type = map(string)
  default = {
    score     = "score"
    threshold = "log"
    strict    = "log"
  }
}

resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "Waf_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # this condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name     = "WAF_Response"
    status   = "403"
    response = "Forbidden"
    content  = "content2"
    request_condition = "WAF_always_false"
  }

  waf {
    prefetch_condition = "Waf_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}

resource "fastly_service_waf_configuration_v1" "waf" {
  waf_id                          = fastly_service_v1.demo.waf[0].waf_id
  http_violation_score_threshold  = 100
}
```

Usage with rules:

```hcl
resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "Waf_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # this condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name     = "WAF_Response"
    status   = "403"
    response = "Forbidden"
    content  = "content2"
    request_condition = "WAF_always_false"
  }

  waf {
    prefetch_condition = "Waf_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}

resource "fastly_service_waf_configuration_v1" "waf" {
  waf_id                          = fastly_service_v1.demo.waf[0].waf_id
  http_violation_score_threshold  = 100

  rule {
    modsec_rule_id = 1010090
    revision       = 1
    status         = "log"
  }
}
```

Usage with rules from datasource:

```hcl
variable "type_status" {
  type = map(string)
  default = {
    score     = "score"
    threshold = "log"
    strict    = "log"
  }
}

resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "Waf_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # this condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name     = "WAF_Response"
    status   = "403"
    response = "Forbidden"
    content  = "content2"
    request_condition = "WAF_always_false"
  }

  waf {
    prefetch_condition = "Waf_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}

data "fastly_waf_rules" "owasp" {
  publishers = ["owasp"]
}

resource "fastly_service_waf_configuration_v1" "waf" {
  waf_id                          = fastly_service_v1.demo.waf[0].waf_id
  http_violation_score_threshold  = 100

  dynamic "rule" {
    for_each = data.fastly_waf_rules.owasp.rules
    content {
      modsec_rule_id = rule.value.modsec_rule_id
      revision       = rule.value.latest_revision_number
      status         = lookup(var.type_status, rule.value.type, "log")
    }
  }
}
```

Usage with rules set individually:

```hcl
variable "type_status" {
  type = map(string)
  default = {
    score     = "score"
    threshold = "log"
    strict    = "log"
  }
}

variable "individual_rules" {
  type = map(string)
  default = {
    1010020 = "block"
  }
}

resource "fastly_service_v1" "demo" {
  name = "demofastly"

  domain {
    name    = "demo.notexample.com"
    comment = "demo"
  }

  backend {
    address = "127.0.0.1"
    name    = "origin1"
    port    = 80
  }

  condition {
    name      = "Waf_Prefetch"
    type      = "PREFETCH"
    statement = "req.backend.is_origin"
  }

  # this condition will always be false
  # adding it to the response object created below
  # prevents Fastly from returning a 403 on all of your traffic
  condition {
    name      = "WAF_always_false"
    statement = "false"
    type      = "REQUEST"
  }

  response_object {
    name     = "WAF_Response"
    status   = "403"
    response = "Forbidden"
    content  = "content2"
    request_condition = "WAF_always_false"
  }

  waf {
    prefetch_condition = "Waf_Prefetch"
    response_object    = "WAF_Response"
  }

  force_destroy = true
}

data "fastly_waf_rules" "owasp" {
  publishers = ["owasp"]
}

resource "fastly_service_waf_configuration_v1" "waf" {
  waf_id                          = fastly_service_v1.demo.waf[0].waf_id
  http_violation_score_threshold  = 202

  dynamic "rule" {
    for_each = data.fastly_waf_rules.owasp.rules
    content {
      modsec_rule_id = rule.value.modsec_rule_id
      revision       = rule.value.latest_revision_number
      status         = lookup(var.individual_rules, rule.value.modsec_rule_id, lookup(var.type_status, rule.value.type, "log"))
    }
  }
}
```


## Argument Reference

The following arguments are supported:

* `waf_id` - (Required) The ID of the Web Application Firewall that the configuration belongs to.
* `allowed_http_versions` - (Optional) Allowed HTTP versions.
* `allowed_methods` - (Optional) A space-separated list of HTTP method names.
* `allowed_request_content_type` - (Optional) Allowed request content types.
* `allowed_request_content_type_charset` - (Required) Allowed request content type charset.
* `arg_length` - (Optional) The maximum number of arguments allowed.
* `arg_name_length` - (Optional) The maximum allowed argument name length.
* `combined_file_sizes` - (Optional) The maximum allowed size of all files.
* `critical_anomaly_score` - (Optional) Score value to add for critical anomalies.
* `crs_validate_utf8_encoding` - (Optional) CRS validate UTF8 encoding.
* `error_anomaly_score` - (Optional) Score value to add for error anomalies.
* `high_risk_country_codes` - (Optional) A space-separated list of country codes in ISO 3166-1 (two-letter) format.
* `http_violation_score_threshold` - (Optional) HTTP violation threshold.
* `inbound_anomaly_score_threshold` - (Optional) Inbound anomaly threshold.
* `lfi_score_threshold` - (Optional) Local file inclusion attack threshold.
* `max_file_size` - (Optional) The maximum allowed file size, in bytes.
* `max_num_args` - (Optional) The maximum number of arguments allowed.
* `notice_anomaly_score` - (Optional) Score value to add for notice anomalies.
* `paranoia_level` - (Optional) The configured paranoia level.
* `php_injection_score_threshold` - (Optional) PHP injection threshold.
* `rce_score_threshold` - (Optional) Remote code execution threshold.
* `restricted_extensions` - (Optional) A space-separated list of allowed file extensions.
* `restricted_headers` - (Optional) A space-separated list of allowed header names.
* `rfi_score_threshold` - (Optional) Remote file inclusion attack threshold.
* `session_fixation_score_threshold` - (Optional) Session fixation attack threshold.
* `sql_injection_score_threshold` - (Optional) SQL injection attack threshold.
* `total_arg_length` - (Optional) The maximum size of argument names and values.
* `warning_anomaly_score` - (Optional) Score value to add for warning anomalies.
* `xss_score_threshold` - (Optional) XSS attack threshold.
* `rule` - (Optional) The Web Application Firewall's active rules.


The `rule` block supports:

* `status` - (Required) The Web Application Firewall rule's status. Allowed values are (log, block and score).
* `modsec_rule_id` - (Required) The Web Application Firewall rule's modsecurity ID.
* `revision` - (Required) The Web Application Firewall rule's revision.


## Import

This is an example of the import command being applied to the resource named `fastly_service_waf_configuration_v1.waf`
The resource ID should be the WAF ID.

```
$ terraform import fastly_service_waf_configuration_v1.waf xxxxxxxxxxxxxxxxxxxx
```

If Terraform is already managing a remote WAF configurations against a resource being imported then the user will be asked to remove it from the existing Terraform state.  
The following is an example of the Terraform state command to remove the resource named `fastly_service_waf_configuration_v1.waf` from the Terraform state file.

```
$ terraform state rm fastly_service_waf_configuration_v1.waf
```
