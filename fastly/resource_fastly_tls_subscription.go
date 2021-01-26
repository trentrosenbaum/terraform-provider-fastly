package fastly

import (
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func resourceFastlyTLSSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceFastlyTLSSubscriptionCreate,
		Read:   resourceFastlyTLSSubscriptionRead,
		Delete: resourceFastlyTLSSubscriptionDelete,
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
			"created_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the subscription was created.",
				Computed:    true,
			},
			"updated_at": {
				Type:        schema.TypeString,
				Description: "Timestamp (GMT) when the subscription was updated.",
				Computed:    true,
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The current state of the subscription. The list of possible states are: `pending`, `processing`, `issued`, and `renewing`.",
				Computed:    true,
			},
		},
	}
}

func resourceFastlyTLSSubscriptionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var configuration *fastly.TLSConfiguration
	if v, ok := d.GetOk("configuration_id"); ok {
		configuration = &fastly.TLSConfiguration{ID: v.(string)}
	}

	var domains []*fastly.TLSDomain
	for _, domain := range d.Get("domains").([]interface{}) {
		domains = append(domains, &fastly.TLSDomain{ID: domain.(string)})
	}

	subscription, err := conn.CreateTLSSubscription(&fastly.CreateTLSSubscriptionInput{
		CertificateAuthority: d.Get("certificate_authority").(string),
		Configuration:        configuration,
		Domain:               domains,
	})
	if err != nil {
		return err
	}

	d.SetId(subscription.ID)

	return resourceFastlyTLSSubscriptionRead(d, meta)
}

func resourceFastlyTLSSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	subscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	err = d.Set("configuration_id", subscription.Configuration.ID)
	if err != nil {
		return err
	}
	err = d.Set("created_at", subscription.CreatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	err = d.Set("updated_at", subscription.UpdatedAt.Format(time.RFC3339))
	if err != nil {
		return err
	}
	err = d.Set("state", subscription.State)
	if err != nil {
		return err
	}

	return nil
}

func resourceFastlyTLSSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeleteTLSSubscription(&fastly.DeleteTLSSubscriptionInput{
		ID: d.Id(),
	})
	return err
}
