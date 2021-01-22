package fastly

import (
	"github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceFastlyTLSConfigurationIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceFastlyTLSConfigurationIDsRead,
		Schema: map[string]*schema.Schema{
			"ids": {
				Type:        schema.TypeList,
				Description: "IDs of available TLS configurations",
				Optional:    true,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceFastlyTLSConfigurationIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	var configurationIDs []string
	cursor := 0
	for {
		list, err := conn.ListCustomTLSConfigurations(&fastly.ListCustomTLSConfigurationsInput{
			PageNumber: cursor,
		})
		if err != nil {
			return err
		}
		if len(list) == 0 {
			break
		}
		cursor += len(list)
		for _, configuration := range list {
			configurationIDs = append(configurationIDs, configuration.ID)
		}
	}

	d.SetId("all") // FIXME: use something more robust when there are some filters
	if err := d.Set("ids", configurationIDs); err != nil {
		return err
	}
	return nil
}
