package fastly

import (
	"fmt"
	"log"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

type ACLServiceAttributeHandler struct {
	*DefaultServiceAttributeHandler
}

func NewServiceACL(sa ServiceMetadata) *ACLServiceAttributeHandler {
	return &ACLServiceAttributeHandler{
		&DefaultServiceAttributeHandler{
			key:             "acl",
			serviceMetadata: sa,
		},
	}
}

func (h *ACLServiceAttributeHandler) Process(d *schema.ResourceData, latestVersion int, conn *gofastly.Client) error {

	o, n := d.GetChange(h.GetKey())
	diff := h.diffSchemaChanges(o, n)

	// Delete removed ACL configurations
	for _, oRaw := range diff.remove {
		opts := h.buildDelete(oRaw, d.Id(), latestVersion)
		err := h.delete(conn, opts)
		if err != nil {
			return err
		}
	}

	// Create new ACL configurations
	for _, vRaw := range diff.add {
		opts := h.buildCreate(vRaw, d.Id(), latestVersion)

		// No point creating a function which just calls gofastly (yet)
		_, err := conn.CreateACL(opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *ACLServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {

	al, err := h.list(conn, d.Id(), s.ActiveVersion.Number)
	if err != nil {
		return err
	}

	alMap := h.flatten(al)
	err = d.Set(h.GetKey(), alMap)
	if err != nil {
		log.Printf("[WARN] Error setting ACLs for (%s): %s", d.Id(), err)
	}

	return nil
}

func (h *ACLServiceAttributeHandler) Register(s *schema.Resource) error {
	s.Schema[h.GetKey()] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				// Required fields
				"name": {
					Type:        schema.TypeString,
					Required:    true,
					Description: "Unique name to refer to this ACL",
				},
				// Optional fields
				"acl_id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Generated acl id",
				},
			},
		},
	}
	return nil
}

func (h *ACLServiceAttributeHandler) buildDelete(oRaw interface{}, serviceID string, serviceVersion int) *gofastly.DeleteACLInput {
	val := oRaw.(map[string]interface{})
	opts := gofastly.DeleteACLInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    val["name"].(string),
	}
	log.Printf("[DEBUG] Fastly ACL removal opts: %#v", opts)
	return &opts
}

func (h *ACLServiceAttributeHandler) delete(conn *gofastly.Client, opts *gofastly.DeleteACLInput) error {
	err := conn.DeleteACL(opts)
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (h *ACLServiceAttributeHandler) buildCreate(oRaw interface{}, serviceID string, serviceVersion int) *gofastly.CreateACLInput {
	val := oRaw.(map[string]interface{})
	opts := gofastly.CreateACLInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    val["name"].(string),
	}
	log.Printf("[DEBUG] Fastly ACL creation opts: %#v", opts)
	return &opts
}

func (h *ACLServiceAttributeHandler) list(conn *gofastly.Client, serviceID string, serviceVersion int) ([]*gofastly.ACL, error) {
	log.Printf("[DEBUG] Refreshing ACLs for (%s)", serviceID)
	// Don't particularly need a buildList function since the variables are explicitly passed
	aclList, err := conn.ListACLs(&gofastly.ListACLsInput{
		Service: serviceID,
		Version: serviceVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("[ERR] Error looking up ACLs for (%s), version (%v): %s", serviceID, serviceVersion, err)
	}
	return aclList, nil
}

func (h *ACLServiceAttributeHandler) flatten(aclList []*gofastly.ACL) []map[string]interface{} {
	var al []map[string]interface{}
	for _, acl := range aclList {
		// Convert VCLs to a map for saving to state.
		vclMap := map[string]interface{}{
			"acl_id": acl.ID,
			"name":   acl.Name,
		}

		// prune any empty values that come from the default string value in structs
		for k, v := range vclMap {
			if v == "" {
				delete(vclMap, k)
			}
		}

		al = append(al, vclMap)
	}

	return al
}
