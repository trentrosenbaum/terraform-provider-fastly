package fastly

import (
	"bytes"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"io/ioutil"
	"text/template"
)

import (
	"strings"
	"testing"
)

const (
	// ManagedByTerraform - default comment/descriptiom
	ManagedByTerraform = "Managed by Terraform"

)

// pgpPublicKey returns a PEM encoded PGP public key suitable for testing.
func pgpPublicKey(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_publickey", t)
}

// privatekey returns a ASN.1 DER encoded key suitable for testing.
func privateKey(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_privatekey", t)
}

// certificate returns a ASN.1 DER encoded certificate for the private key suitable for testing.
func certificate(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_certificate", t)
}

// caCert returns a CA certificate suitable for testing
func caCert(t *testing.T) string {
	return readTestFile("./test_fixtures/fastly_test_cacert", t)
}

func readTestFile(filename string, t *testing.T) string {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		t.Fatalf("Cannot load key file: %s", filename)
	}
	return string(contents)
}

// escapePercentSign uses Terraform's escape syntax (i.e., repeating characters)
// to properly escape percent signs (i.e., '%').
//
// There are two significant places where '%' can show up:
// 1. Before a left curly brace (i.e., '{').
// 2. Not before a left curly brace.
//
// In case #1, we have to double escape so that Terraform does not interpret Fastly's
// configuration values as its own (e.g., https://docs.fastly.com/en/guides/custom-log-formats).
//
// In case #2, we only have to single escape.
//
// Refer to the Terraform documentation on string literals for more details:
// https://www.terraform.io/docs/configuration/expressions.html#string-literals
func escapePercentSign(s string) string {
	escapeSign := strings.ReplaceAll(s, "%", "%%")
	escapeTemplateSequence := strings.ReplaceAll(escapeSign, "%%{", "%%%%{")

	return escapeTemplateSequence
}

func TestEscapePercentSign(t *testing.T) {
	for _, testcase := range []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "string no percent signs should change nothing",
			input: "abc 123",
			want:  "abc 123",
		},
		{
			name:  "one percent sign should return two percent signs",
			input: "%",
			want:  "%%",
		},
		{
			name:  "one percent sign mid-string should return two percent signs in the same place",
			input: "abc%123",
			want:  "abc%%123",
		},
		{
			name:  "one percent sign before left curly brace should return four percent signs then curly brace",
			input: "%{",
			want:  "%%%%{",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			got := escapePercentSign(testcase.input)

			if got != testcase.want {
				t.Errorf("escapePercentSign(%q): \n\tgot: '%+v'\n\twant: '%+v'", testcase.input, got, testcase.want)
			}
		})
	}
}

func appendNewLine(s string) string {
	return s + "\n"
}

func makeTestDomainName() string {
	return fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10))
}

func makeTestServiceName() string {
	return fmt.Sprintf("tf_test_%s", acctest.RandString(10))
}

func makeTestBlockName(block string) string {
	return fmt.Sprintf("tf_test_%s_%s", block, acctest.RandString(10))
}

func makeTestServiceComment() string {
	return fmt.Sprintf("tf-test-%s", acctest.RandString(10))
}

func testGetResourceTemplate(filename string, data map[string] interface{}) string {
	var filepath = fmt.Sprintf("resource_templates/%v.tmpl", filename)
	t := template.Must(template.ParseFiles(filepath))

	var buf bytes.Buffer
	t.Execute(&buf, data)
	return buf.String()
}