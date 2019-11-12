package fastly

import (
	"log"
	"strconv"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

// WAFSchema the WAF block schema
var WAFSchema = &schema.Schema{
	Type:     schema.TypeList,
	Optional: true,
	MaxItems: 1,
	Elem: &schema.Resource{
		Schema: map[string]*schema.Schema{
			"response_object": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The web firewall's response object",
			},
			"prefetch_condition": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The web firewall's prefetch condition",
			},
			"waf_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The web firewall id",
			},
		},
	},
}

func processWAF(d *schema.ResourceData, conn *gofastly.Client, v int) error {

	serviceID := d.Id()
	serviceVersion := strconv.Itoa(v)
	oldWAFVal, newWAFVal := d.GetChange("waf")

	if len(oldWAFVal.([]interface{})) > 0 && len(newWAFVal.([]interface{})) > 0 {
		wf := newWAFVal.([]interface{})[0].(map[string]interface{})
		opts, err := buildWAF(wf, serviceID, serviceVersion)
		if err != nil {
			log.Printf("[DEBUG] Error building WAF: %s", err)
			return err
		}

		log.Printf("[DEBUG] Fastly WAF update opts: %#v", opts)

		// check if WAF exists first
		if !wAFExists(conn, gofastly.GetWAFInput{
			Version: serviceVersion,
			Service: serviceID,
			ID:      wf["waf_id"].(string),
		}) {
			log.Printf("[WARN] WAF not found, creating one with update opts: %#v", opts)
			if err := createWAF(wf, conn, opts); err != nil {
				return err
			}
		}
		_, err = conn.UpdateWAF(opts)
		if err != nil {
			return err
		}

	} else if len(newWAFVal.([]interface{})) > 0 {
		wf := newWAFVal.([]interface{})[0].(map[string]interface{})
		opts, err := buildWAF(wf, serviceID, serviceVersion)
		if err != nil {
			log.Printf("[DEBUG] Error building WAF: %s", err)
			return err
		}

		log.Printf("[DEBUG] Fastly WAF addition opts: %#v", opts)
		if err := createWAF(wf, conn, opts); err != nil {
			return err
		}

	} else if len(oldWAFVal.([]interface{})) > 0 {

		log.Printf("[DEBUG] deleting WAF")
		df := oldWAFVal.([]interface{})[0].(map[string]interface{})

		opts := gofastly.DeleteWAFInput{
			Version: serviceVersion,
			ID:      df["waf_id"].(string),
		}

		log.Printf("[DEBUG] Fastly WAF Removal opts: %#v", opts)
		err := conn.DeleteWAF(&opts)
		if errRes, ok := err.(*gofastly.HTTPError); ok {
			if errRes.StatusCode != 404 {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return nil
}

func createWAF(df map[string]interface{}, conn *gofastly.Client, i *gofastly.WAFInput) error {

	log.Printf("[DEBUG] Fastly WAF Addition opts: %#v", i)
	w, err := conn.CreateWAF(i)
	if err != nil {
		return err
	}
	df["waf_id"] = w.ID
	return nil
}

func wAFExists(conn *gofastly.Client, i gofastly.GetWAFInput) bool {

	_, err := conn.GetWAF(&i)
	if err != nil {
		return false
	}
	return true
}

func flattenWAFs(wafList []*gofastly.WAF) []map[string]interface{} {

	var wl []map[string]interface{}
	if len(wafList) == 0 {
		return wl
	}

	w := wafList[0]
	WAFMapString := map[string]interface{}{
		"waf_id":             w.ID,
		"response_object":    w.Response,
		"prefetch_condition": w.PrefetchCondition,
	}

	// prune any empty values that come from the default string value in structs
	for k, v := range WAFMapString {
		if v == "" {
			delete(WAFMapString, k)
		}
	}
	return append(wl, WAFMapString)
}

func buildWAF(WAFMap interface{}, serviceID string, ServiceVersion string) (*gofastly.WAFInput, error) {
	df := WAFMap.(map[string]interface{})
	opts := gofastly.WAFInput{
		Service:           serviceID,
		Version:           ServiceVersion,
		ID:                df["waf_id"].(string),
		PrefetchCondition: df["prefetch_condition"].(string),
		Response:          df["response_object"].(string),
	}
	return &opts, nil
}
