---
layout: "fastly"
page_title: "Fastly: fastly_tls_domain"
sidebar_current: "docs-fastly-datasource-tls_domain"
description: |-
Get IDs of activations, certificates and subscriptions associated with a domain.
---

# fastly_tls_domain

Use this data source to get the IDs of activations, certificates and subscriptions associated with a domain.

## Example Usage

```hcl
data "fastly_tls_domain" "example" {
  domain = "terraform.fastly.example"
}
```

## Argument Reference

This data source has no arguments.

* `domain` - (Required) Domain name

## Attribute Reference

* `tls_activation_ids` - IDs of the activations associated with the domain
* `tls_certificate_ids` - IDs of the certificates associated with the domain
* `tls_subscription_ids` - IDs of the subscriptions associated with the domain