package fastly

import (
	"fmt"
	"time"

	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func resourceFastlyTLSSubscription() *schema.Resource {
	return &schema.Resource{
		Create: resourceFastlyTLSSubscriptionCreate,
		Read:   resourceFastlyTLSSubscriptionRead,
		Delete: resourceFastlyTLSSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"domains": {
				Type:        schema.TypeSet,
				Description: "List of domains on which to enable TLS.",
				Required:    true,
				ForceNew:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				MinItems:    1,
				Set:         schema.HashString,
			},
			"certificate_authority": {
				Type:         schema.TypeString,
				Description:  "The entity that issues and certifies the TLS certificates for your subscription. Valid values are `lets-encrypt` or `globalsign`.",
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"lets-encrypt", "globalsign"}, false),
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
			"managed_dns_challenge": {
				Type:        schema.TypeMap,
				Description: "The details required to configure DNS to respond to ACME DNS challenge in order to verify domain ownership.",
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"managed_http_challenges": {
				Type:        schema.TypeSet,
				Description: "A list of options for configuring DNS to respond to ACME HTTP challenge in order to verify domain ownership. Best accessed through a `for` expression to filter the relevant record.",
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"record_type": {
							Type:        schema.TypeString,
							Description: "The name of the DNS record to add. For example `example.com`.",
							Computed:    true,
						},
						"record_name": {
							Type:        schema.TypeString,
							Description: "The type of DNS record to add, e.g. `A`, or `CNAME`.",
							Computed:    true,
						},
						"record_values": {
							Type:        schema.TypeSet,
							Description: "A list with the value(s) to which the DNS record should point.",
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
						},
					},
				},
				Set: authorisationChallengesHash,
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
	for _, domain := range d.Get("domains").(*schema.Set).List() {
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

	include := "tls_authorizations"
	subscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
		ID:      d.Id(),
		Include: &include,
	})
	if err != nil {
		return err
	}

	var domains []string
	for _, domain := range subscription.TLSDomains {
		domains = append(domains, domain.ID)
	}

	var managedHTTPChallenges []map[string]interface{}
	var managedDNSChallenge map[string]string
	for _, challenge := range subscription.Authorizations[0].Challenges {
		if challenge.Type == "managed-dns" {
			if len(challenge.Values) < 1 {
				return fmt.Errorf("Fastly API returned no record values for Managed DNS Challenge")
			}

			managedDNSChallenge = map[string]string{
				"record_type":  challenge.RecordType,
				"record_name":  challenge.RecordName,
				"record_value": challenge.Values[0],
			}
		} else {
			managedHTTPChallenges = append(managedHTTPChallenges, map[string]interface{}{
				"record_type":   challenge.RecordType,
				"record_name":   challenge.RecordName,
				"record_values": challenge.Values,
			})
		}
	}

	err = d.Set("domains", domains)
	if err != nil {
		return err
	}
	err = d.Set("certificate_authority", subscription.CertificateAuthority)
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
	err = d.Set("managed_dns_challenge", managedDNSChallenge)
	if err != nil {
		return err
	}
	err = d.Set("managed_http_challenges", managedHTTPChallenges)
	if err != nil {
		return err
	}

	return nil
}

func resourceFastlyTLSSubscriptionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	subscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	// Delete any associated TLS activations using this subscription
	if subscription.Certificates != nil && len(subscription.Certificates) > 0 {
		certificateID := subscription.Certificates[0].ID

		activations, err := conn.ListTLSActivations(&fastly.ListTLSActivationsInput{
			FilterTLSCertificateID: certificateID,
		})
		if err != nil {
			return err
		}

		for _, activation := range activations {
			if activation.Certificate.ID != certificateID {
				return fmt.Errorf("Fastly API returned a TLS activation for a different subscription or certificate")
			}

			err := conn.DeleteTLSActivation(&fastly.DeleteTLSActivationInput{
				ID: activation.ID,
			})
			if err != nil {
				return err
			}
		}
	}

	err = conn.DeleteTLSSubscription(&fastly.DeleteTLSSubscriptionInput{
		ID: d.Id(),
	})
	return err
}

func authorisationChallengesHash(value interface{}) int {
	m, ok := value.(map[string]interface{})
	if !ok {
		return 0
	}

	recordType, ok := m["record_type"].(string)
	if !ok {
		return 0
	}

	recordName, ok := m["record_name"].(string)
	if ok {
		return hashcode.String(fmt.Sprintf("%s_%s", recordType, recordName))
	}

	return 0
}
