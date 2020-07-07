package fastly

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

// DefaultServiceAttributeHandler provides a base implementation for ServiceAttributeDefinition.
type DefaultServiceAttributeHandler struct {
	key             string
	serviceMetadata ServiceMetadata
}

// GetKey is provided since most attributes will just use their private "key" for interacting with the service.
func (h *DefaultServiceAttributeHandler) GetKey() string {
	return h.key
}

// GetServiceType is provided to allow internal methods to get the service Type.
func (h *DefaultServiceAttributeHandler) GetServiceMetadata() ServiceMetadata {
	return h.serviceMetadata
}

// HasChange returns whether the state of the attribute has changed against Terraform stored state.
// This implements a method from ServiceAttributeDefinition.
func (h *DefaultServiceAttributeHandler) HasChange(d *schema.ResourceData) bool {
	return d.HasChange(h.key)
}

// MustProcess returns whether we must process the resource (usually HasChange==true but allowing exceptions).
// For example: at present, the settings attributeHandler (block_fastly_service_v1_settings.go) must process when
// default_ttl==0 and it is the initialVersion - as well as when default_ttl or default_host have changed.
// This implements a method from ServiceAttributeDefinition.
func (h *DefaultServiceAttributeHandler) MustProcess(d *schema.ResourceData, initialVersion bool) bool {
	return h.HasChange(d)
}

// VCLLoggingAttributes is a container for vcl-only attributes in logging blocks.
type VCLLoggingAttributes struct {
	format            string
	formatVersion     uint
	placement         string
	responseCondition string
}

// NewVCLLoggingAttributes provides default values to Compute services for VCL only logging attributes.
func (h *DefaultServiceAttributeHandler) getVCLLoggingAttributes(data map[string]interface{}) VCLLoggingAttributes {
	var vla = VCLLoggingAttributes{
		placement: "none",
	}
	if h.GetServiceMetadata().serviceType == ServiceTypeVCL {
		if val, ok := data["format"]; ok {
			vla.format = val.(string)
		}
		if val, ok := data["format_version"]; ok {
			vla.formatVersion = uint(val.(int))
		}
		if val, ok := data["placement"]; ok {
			vla.placement = val.(string)
		}
		if val, ok := data["response_condition"]; ok {
			vla.responseCondition = val.(string)
		}
	}
	return vla
}

// DiffSchemaChangesResult is a container for returning the result of a schema old/new diff.
type DiffSchemaChangesResult struct {
	add    []interface{}
	remove []interface{}
}

func (h *DefaultServiceAttributeHandler) diffSchemaChanges(o interface{}, n interface{}) DiffSchemaChangesResult {

	if o == nil {
		o = new(schema.Set)
	}
	if n == nil {
		n = new(schema.Set)
	}

	os := o.(*schema.Set)
	ns := n.(*schema.Set)

	return DiffSchemaChangesResult{
		ns.Difference(os).List(),
		os.Difference(ns).List(),
	}
}
