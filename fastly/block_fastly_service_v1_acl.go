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
	return ProcessServiceAttribute(h, d, latestVersion, conn)
}

func (h *ACLServiceAttributeHandler) Read(d *schema.ResourceData, s *gofastly.ServiceDetail, conn *gofastly.Client) error {
	return ReadServiceAttribute(h, d, s, conn)
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

func (h *ACLServiceAttributeHandler) buildDelete(oRaw interface{}, serviceID string, serviceVersion int) interface{} {
	val := oRaw.(map[string]interface{})
	opts := gofastly.DeleteACLInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    val["name"].(string),
	}
	log.Printf("[DEBUG] Fastly ACL removal opts: %#v", opts)
	return &opts
}

func (h *ACLServiceAttributeHandler) delete(conn *gofastly.Client, opts interface{}) error {
	err := conn.DeleteACL(opts.(*gofastly.DeleteACLInput))
	if errRes, ok := err.(*gofastly.HTTPError); ok {
		if errRes.StatusCode != 404 {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (h *ACLServiceAttributeHandler) buildCreate(oRaw interface{}, serviceID string, serviceVersion int) interface{} {
	val := oRaw.(map[string]interface{})
	opts := gofastly.CreateACLInput{
		Service: serviceID,
		Version: serviceVersion,
		Name:    val["name"].(string),
	}
	log.Printf("[DEBUG] Fastly ACL creation opts: %#v", opts)
	return &opts
}

func (h *ACLServiceAttributeHandler) create(conn *gofastly.Client, opts interface{}) error {
	_, err := conn.CreateACL(opts.(*gofastly.CreateACLInput))
	return err
}

func (h *ACLServiceAttributeHandler) list(conn *gofastly.Client, serviceID string, serviceVersion int) ([]interface{}, error) {
	log.Printf("[DEBUG] Refreshing ACLs for (%s)", serviceID)
	// Don't particularly need a buildList function since the variables are explicitly passed
	aclList, err := conn.ListACLs(&gofastly.ListACLsInput{
		Service: serviceID,
		Version: serviceVersion,
	})
	if err != nil {
		return nil, fmt.Errorf("[ERR] Error looking up ACLs for (%s), version (%v): %s", serviceID, serviceVersion, err)
	}
	return h.listToGeneric(aclList), nil
}

func (h *ACLServiceAttributeHandler) flatten(aclList []interface{}) []map[string]interface{} {
	var al []map[string]interface{}
	for _, acl := range aclList {
		ao := acl.(*gofastly.ACL)
		al = append(al, h.pruneMap(map[string]interface{}{
			"acl_id": ao.ID,
			"name":   ao.Name,
		}))
	}
	return al
}

func (h *ACLServiceAttributeHandler) listToGeneric(src []*gofastly.ACL) []interface{} {
	genericList := make([]interface{}, len(src))
	for i := range src {
		genericList[i] = src[i]
	}
	return genericList
}
