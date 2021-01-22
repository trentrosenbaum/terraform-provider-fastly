package fastly

import (
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSCertificateIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSCertificateIDsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Description: "IDs of custom TLS certificates",
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSCertificateIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var certificateIDs []string
	pageNumber := 1
	for {
		list, err := conn.ListCustomTLSCertificates(&fastly.ListCustomTLSCertificatesInput{
			PageNumber: pageNumber,
		})
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		pageNumber++

		for _, certificate := range list {
			certificateIDs = append(certificateIDs, certificate.ID)
		}
	}

	d.SetId("all") // FIXME: use something more robust when there are some filters
	err := d.Set("ids", certificateIDs)
	if err != nil {
		return err
	}

	return nil
}
