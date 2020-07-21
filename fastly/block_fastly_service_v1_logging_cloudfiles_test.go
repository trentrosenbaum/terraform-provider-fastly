package fastly

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	gofastly "github.com/fastly/go-fastly/fastly"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// SuperTestCloudFilesData provides a data structure for multiple CloudFiles logging blocks.
// This uses the existing gofastly structure for CloudFiles
type SuperTestCloudFilesData struct {
	CloudFiles []*gofastly.Cloudfiles
}

// SuperTestComponentCloudFiles mixes-in the shared SuperTestComponent functions with SuperTestCloudFilesData.
type SuperTestComponentCloudFiles struct {
	*SuperTestComponentShared
	*SuperTestCloudFilesData
}

// NewSuperTestComponentCloudFiles is the constructor for a SuperTestComponentCloudFiles.
// It injects the service metadata and creates a single random gofastly.Cloudfiles block.
func NewSuperTestComponentCloudFiles(service *SuperTestService) *SuperTestComponentCloudFiles {
	var c SuperTestComponentCloudFiles
	c.SuperTestComponentShared = &SuperTestComponentShared{
		service,
	}
	c.SuperTestCloudFilesData = &SuperTestCloudFilesData{
		[]*gofastly.Cloudfiles{
			c.randomCloudFiles(),
		},
	}
	return &c
}

// randomCloudFiles creates a random gofastly.Cloudfiles block.
func (c *SuperTestComponentCloudFiles) randomCloudFiles() *gofastly.Cloudfiles {
	return &gofastly.Cloudfiles{
		Name:              fmt.Sprintf("tf-test-logging-cloudfiles-%s", acctest.RandString(5)),
		BucketName:        fmt.Sprintf("tf-test-logging-cloudfiles-bucket-%s", acctest.RandString(5)),
		User:              fmt.Sprintf("user_%s", acctest.RandString(5)),
		AccessKey:         fmt.Sprintf("secret_%s", acctest.RandString(5)),
		PublicKey:         pgpPublicKey(nil),
		GzipLevel:         c.randomGzipLevel(),
		MessageType:       c.randomChoiceString([]string{"classic", "loggly", "logplex", "blank"}),
		Path:              fmt.Sprintf("%s/", acctest.RandString(5)),
		Region:            c.randomChoiceString([]string{"ORD", "LON", "SYD", "DFW", "IAD", "HKG"}),
		Period:            c.randomVarianceUint(3600, 1000),
		TimestampFormat:   c.randomTimestampFormat(),
		Format:            c.randomFormat(),
		FormatVersion:     c.randomFormatVersion(),
		ResponseCondition: "", // ToDo: Add Condition to V1 test and reference in this block
		Placement:         c.randomPlacement(),
	}
}

// AddRandom fulfils the SuperTestComponent interface.
// It appends a random gofastly.Cloudfiles block to the list.
func (c *SuperTestComponentCloudFiles) AddRandom() error {
	c.CloudFiles = append(c.CloudFiles, c.randomCloudFiles())
	return nil
}

// UpdateRandom fulfils the SuperTestComponent interface.
// It updates a randomly selected gofastly.Cloudfiles block to a new random set of values.
func (c *SuperTestComponentCloudFiles) UpdateRandom() error {
	var randomIndex = c.randomIndex()
	c.CloudFiles[randomIndex] = c.randomCloudFiles()
	return nil
}

// DeleteRandom fulfils the SuperTestComponent interface.
// It deletes a random gofastly.Cloudfiles block.
func (c *SuperTestComponentCloudFiles) DeleteRandom() error {
	var numLoggingCloudFiles = len(c.CloudFiles)
	if numLoggingCloudFiles < 2 {
		return fmt.Errorf("cannot delete random cloudfiles logging block only have %v", numLoggingCloudFiles)
	}
	var randomIndex = c.randomIndex()
	c.CloudFiles = append(c.CloudFiles[:randomIndex], c.CloudFiles[randomIndex+1:]...)
	return nil
}

// randomIndex selects an index at random from the list of gofastly.Cloudfiles blocks.
func (c *SuperTestComponentCloudFiles) randomIndex() int {
	return rand.Intn(len(c.CloudFiles))
}

// GetTestCheckFunc fulfils the SuperTestComponent interface.
// It composes a series of tests to verify this component of a provisioned resource against
// the data used to provision it.
func (c *SuperTestComponentCloudFiles) GetTestCheckFunc(serviceDetail *gofastly.ServiceDetail) []resource.TestCheckFunc {
	var numLoggingCloudfiles = len(c.CloudFiles)
	var r = []resource.TestCheckFunc{
		resource.TestCheckResourceAttr(c.Service.ResourceId, "logging_cloudfiles.#", strconv.Itoa(numLoggingCloudfiles)),
		c.testState(),
		c.testApi(serviceDetail),
	}
	return r
}

// testState checks a provisioned resource CloudFiles logging *state* against the data used to provision it.
func (c *SuperTestComponentCloudFiles) testState() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var sourceList = flattenCloudfiles(c.CloudFiles)
		var targetList, err = c.getStateTypeSetBlocks(s, "logging_cloudfiles")
		if err != nil {
			return err
		}

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("state cloudfiles_logging mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testApi checks a provisioned resource CloudFiles logging at the *api* against the data used to provision it.
func (c *SuperTestComponentCloudFiles) testApi(serviceDetail *gofastly.ServiceDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*FastlyClient).conn
		apiList, err := conn.ListCloudfiles(&gofastly.ListCloudfilesInput{
			Service: serviceDetail.ID,
			Version: serviceDetail.Version.Number,
		})
		if err != nil {
			return err
		}

		var sourceList = flattenCloudfiles(c.CloudFiles)
		var targetList = flattenCloudfiles(apiList)

		if !c.testEquivalent(sourceList, targetList) {
			return fmt.Errorf("api cloudfiles mismatch, expected: %#v, got: %#v", sourceList, targetList)
		}

		return nil
	}
}

// testEquivalent compares two flattened data Domains
func (c *SuperTestComponentCloudFiles) testEquivalent(ms1 []map[string]interface{}, ms2 []map[string]interface{}) bool {
	expected := len(ms1)
	for _, m1 := range ms1 {
		for _, m2 := range ms2 {

			// Flatten removes empty keys, so must run this to ensure equivalence
			c.removeEmptyKeys(&m1)
			c.removeEmptyKeys(&m2)

			ok := c.compareKey(m1, m2, "name") &&
				c.compareKey(m1, m2, "bucket_name") &&
				c.compareKey(m1, m2, "user") &&
				c.compareKey(m1, m2, "access_key") &&
				c.compareKey(m1, m2, "public_key") &&
				c.compareKey(m1, m2, "gzip_level") &&
				c.compareKey(m1, m2, "message_type") &&
				c.compareKey(m1, m2, "path") &&
				c.compareKey(m1, m2, "region") &&
				c.compareKey(m1, m2, "period") &&
				c.compareKey(m1, m2, "timestamp_format")

			if ok && c.Service.Type == "vcl" {
				ok = c.compareKey(m1, m2, "format") &&
					c.compareKey(m1, m2, "format_version") &&
					c.compareKey(m1, m2, "response_condition") &&
					c.compareKey(m1, m2, "placement")
			}

			if ok {
				expected--
			}
		}
	}
	return expected <= 0
}

func (c *SuperTestComponentCloudFiles) removeEmptyKeys(m *map[string]interface{}) {
	for k, v := range *m {
		if v == "" {
			delete(*m, k)
		}
	}
}

func TestResourceFastlyFlattenCloudfiles(t *testing.T) {
	cases := []struct {
		remote []*gofastly.Cloudfiles
		local  []map[string]interface{}
	}{
		{
			remote: []*gofastly.Cloudfiles{
				{
					Version:           1,
					Name:              "cloudfiles-endpoint",
					BucketName:        "bucket",
					User:              "user",
					AccessKey:         "secret",
					PublicKey:         pgpPublicKey(t),
					Format:            "%h %l %u %t \"%r\" %>s %b",
					FormatVersion:     2,
					GzipLevel:         0,
					MessageType:       "classic",
					Path:              "/",
					Region:            "ORD",
					Period:            3600,
					Placement:         "none",
					ResponseCondition: "response_condition",
					TimestampFormat:   "%Y-%m-%dT%H:%M:%S.000",
				},
			},
			local: []map[string]interface{}{
				{
					"name":               "cloudfiles-endpoint",
					"bucket_name":        "bucket",
					"user":               "user",
					"access_key":         "secret",
					"public_key":         pgpPublicKey(t),
					"format":             "%h %l %u %t \"%r\" %>s %b",
					"format_version":     uint(2),
					"gzip_level":         uint(0),
					"message_type":       "classic",
					"path":               "/",
					"region":             "ORD",
					"period":             uint(3600),
					"placement":          "none",
					"response_condition": "response_condition",
					"timestamp_format":   "%Y-%m-%dT%H:%M:%S.000",
				},
			},
		},
	}

	for _, c := range cases {
		out := flattenCloudfiles(c.remote)
		if diff := cmp.Diff(out, c.local); diff != "" {
			t.Fatalf("Error matching: %s", diff)
		}
	}
}
