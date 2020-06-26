package fastly

import (
	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// ServiceAttributeHandlerDefinition provides an interface for service attributes.
// We compose a service resource out of attribute objects to allow us to construct both the VCL and Wasm service
// resources from common components.
type ServiceAttributeHandlerDefinition interface {

	// Register add the attribute to the resource schema.
	Register(s *schema.Resource, serviceType string) error

	// Read refreshes the attribute state against the Fastly API.
	Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error

	// Process creates or updates the attribute against the Fastly API.
	Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error

	// HasChange returns whether the state of the attribute has changed against Terraform stored state.
	HasChange(d *schema.ResourceData) bool

	// MustProcess returns whether we must process the resource (usually HasChange==true but allowing exceptions).
	// For example: at present, the settings attributeHandler (block_fastly_service_v1_settings.go) must process when
	// default_ttl==0 and it is the initialVersion - as well as when default_ttl or default_host have changed.
	MustProcess(d *schema.ResourceData, initialVersion bool) bool
}

// DefaultServiceAttributeHandler provides a base implementation for ServiceAttributeHandlerDefinition.
type DefaultServiceAttributeHandler struct {
	key         string
	serviceType string
}

// See interface definition for comments.
func (h *DefaultServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChange(h.key)
}

// See interface definition for comments.
func (h *DefaultServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return h.HasChange(d)
}

// GetKey is provided since most attributes will just use their private "key" for interacting with the service.
// Not in the interface since this shouldn't be used publicly
func (h *DefaultServiceAttributeHandler) GetKey() string {
	return h.key
}

// GetServiceType is provided to allow differential processing where the owning service is Wasm or VCL.
// Not in the interface since this shouldn't be used publicly
func (h *DefaultServiceAttributeHandler) GetServiceType() string {
	return h.serviceType
}

// OptionalMapKeyToString returns a default if the key is not found in the map
// This is used for attributes which are now optional in a Wasm service
// Not in the interface since this shouldn't be used publicly
func (h *DefaultServiceAttributeHandler) OptionalMapKeyToString(m map[string]interface{}, k string, d string) string {
	v, ok := m[k]
	if ok {
		return v.(string)
	} else {
		return d
	}
}

// OptionalMapKeyToUInt returns a default if the key is not found in the map
// This is used for attributes which are now optional in a Wasm service
// Not in the interface since this shouldn't be used publicly
func (h *DefaultServiceAttributeHandler) OptionalMapKeyToUInt(m map[string]interface{}, k string, d uint) uint {
	v, ok := m[k]
	if ok {
		return uint(v.(int))
	} else {
		return d
	}
}

type VCLLoggingAttributes struct {
	format            string
	formatVersion     uint
	placement         string
	responseCondition string
}

// NewSomething create new instance of Something
func NewVCLLoggingAttributes() VCLLoggingAttributes {
	vla := VCLLoggingAttributes{}
	vla.format = ""
	vla.formatVersion = 0
	vla.placement = "none"
	vla.responseCondition = ""
	return vla
}
