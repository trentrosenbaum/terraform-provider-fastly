package fastly

import (
	"fmt"
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"time"
)

func dataSourceFastlyTLSSubscription() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSSubscriptionRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of TLS subscription.",
				ConflictsWith: []string{"configuration_id", "domains", "certificate_authority"},
			},
			"configuration_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of TLS configuration used to terminate TLS traffic.",
				ConflictsWith: []string{"id"},
			},
			"domains": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "List of domains on which to enable TLS.",
				ConflictsWith: []string{"id"},
			},
			"certificate_authority": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The entity that issues and certifies the TLS certificates for the subscription.",
				ConflictsWith: []string{"id"},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (GMT) when subscription was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp (GMT) when subscription was last updated.",
			},
			"state": {
				Type:        schema.TypeString,
				Description: "The current state of the subscription. The list of possible states are: `pending`, `processing`, `issued`, and `renewing`.",
				Computed:    true,
			},
		},
	}
}

func dataSourceFastlyTLSSubscriptionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var subscription *fastly.TLSSubscription

	if v, ok := d.GetOk("id"); ok {
		foundSubscription, err := conn.GetTLSSubscription(&fastly.GetTLSSubscriptionInput{
			ID: v.(string),
		})
		if err != nil {
			return err
		}
		subscription = foundSubscription
	} else {
		filters := getTLSSubscriptionFilters(d)
		subscriptions, err := listTLSSubscriptions(conn, filters...)
		if err != nil {
			return err
		}

		if len(subscriptions) == 0 {
			return fmt.Errorf("Your query returned no results. Please change your search criteria and try again")
		}

		if len(subscriptions) > 1 {
			return fmt.Errorf("Your query returned more than one result. Please change to a more specific search criteria")
		}

		subscription = subscriptions[0]
	}

	return dataSourceFastlyTLSSubscriptionSetAttributes(subscription, d)
}

type TLSSubscriptionPredicate func(*fastly.TLSSubscription) bool

func getTLSSubscriptionFilters(d *schema.ResourceData) []TLSSubscriptionPredicate {
	var filters []TLSSubscriptionPredicate

	if v, ok := d.GetOk("configuration_id"); ok {
		filters = append(filters, func(s *fastly.TLSSubscription) bool {
			return s.Configuration.ID == v.(string)
		})
	}
	if v, ok := d.GetOk("domains"); ok {
		domainsToMatch := v.(*schema.Set).List()
		filters = append(filters, func(s *fastly.TLSSubscription) bool {
			// Pull domain strings out of struct slice
			var foundDomains []string
			for _, domain := range s.TLSDomains {
				foundDomains = append(foundDomains, domain.ID)
			}

			return containsSubSet(foundDomains, domainsToMatch)
		})
	}
	if v, ok := d.GetOk("certificate_authority"); ok {
		filters = append(filters, func(s *fastly.TLSSubscription) bool {
			return s.CertificateAuthority == v.(string)
		})
	}

	return filters
}

func listTLSSubscriptions(conn *fastly.Client, filters ...TLSSubscriptionPredicate) ([]*fastly.TLSSubscription, error) {
	var subscriptions []*fastly.TLSSubscription
	pageNumber := 1
	for {
		list, err := conn.ListTLSSubscriptions(&fastly.ListTLSSubscriptionsInput{
			PageNumber: pageNumber,
			PageSize:   10,
		})
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		for _, subscription := range list {
			if filterTLSSubscriptions(subscription, filters) {
				subscriptions = append(subscriptions, subscription)
			}
		}
	}

	return subscriptions, nil
}

func dataSourceFastlyTLSSubscriptionSetAttributes(subscription *fastly.TLSSubscription, d *schema.ResourceData) error {
	d.SetId(subscription.ID)

	var domains []string
	for _, domain := range subscription.TLSDomains {
		domains = append(domains, domain.ID)
	}

	err := d.Set("configuration_id", subscription.Configuration.ID)
	if err != nil {
		return err
	}
	err = d.Set("domains", domains)
	if err != nil {
		return err
	}
	err = d.Set("certificate_authority", subscription.CertificateAuthority)
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

func filterTLSSubscriptions(subscription *fastly.TLSSubscription, filters []TLSSubscriptionPredicate) bool {
	for _, f := range filters {
		if !f(subscription) {
			return false
		}
	}
	return true
}
