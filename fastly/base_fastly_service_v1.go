package fastly

import (
	"github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// SERVICE ATTRIBUTE

type ServiceAttributeDefinition interface {
	GetKey() string
	Read(d *schema.ResourceData, s *fastly.ServiceDetail, conn *fastly.Client) error
	Process(d *schema.ResourceData, latestVersion int, conn *fastly.Client) error
	Register(d *schema.Resource) error
}

type DefaultServiceAttributeHandler struct {
	key    string
}

func (h *DefaultServiceAttributeHandler) GetKey() string {
	return h.key
}

