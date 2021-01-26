package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceFastlyTLSSubscription() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeList,
				Description: "List of domains on which to enable TLS.",
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"certificate_authority": {
				Type:        schema.TypeString,
				Description: "The entity that issues and certifies the TLS certificates for your subscription. Valid values are `lets-encrypt` or `globalsign`.",
				Required:    true,
				ForceNew:    true,
			},
			"configuration_id": {
				Type:        schema.TypeString,
				Description: "The ID of the set of TLS configuration options that apply to the enabled domains on this subscription.",
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
			},
		},
	}
}
