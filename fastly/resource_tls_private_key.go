package fastly

import (
	gofastly "github.com/fastly/go-fastly/v2/fastly"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceTLSPrivateKeyV1() *schema.Resource {
	return &schema.Resource{
		Create: resourceTLSPrivateKeyV1Create,
		Read:   resourceTlSPrivateKeyV1Read,
		Delete: resourceTLSPrivateKeyV1Delete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key_pem": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Private key in PEM format.",
				Sensitive:   true,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Customisable name of the private key.",
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"key_length": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"key_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replace": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether Fastly recommends replacing this private key.",
			},
			"public_key_sha1": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTLSPrivateKeyV1Create(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	privateKey, err := conn.CreatePrivateKey(&gofastly.CreatePrivateKeyInput{
		Key:  d.Get("key_pem").(string),
		Name: d.Get("name").(string),
	})
	if err != nil {
		return err
	}

	d.SetId(privateKey.ID)

	return resourceTlSPrivateKeyV1Read(d, meta)
}

func resourceTlSPrivateKeyV1Read(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	privateKey, err := conn.GetPrivateKey(&gofastly.GetPrivateKeyInput{
		ID: d.Id(),
	})
	if err != nil {
		return err
	}

	err = d.Set("name", privateKey.Name)
	err = d.Set("created_at", privateKey.CreatedAt.String())
	err = d.Set("key_length", privateKey.KeyLength)
	err = d.Set("key_type", privateKey.KeyType)
	err = d.Set("replace", privateKey.Replace)
	err = d.Set("public_key_sha1", privateKey.PublicKeySHA1)

	return err
}

func resourceTLSPrivateKeyV1Delete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*FastlyClient).conn

	err := conn.DeletePrivateKey(&gofastly.DeletePrivateKeyInput{
		ID: d.Id(),
	})

	return err
}
