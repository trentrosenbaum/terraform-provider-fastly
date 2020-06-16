package fastly

import (
	"io/ioutil"
)

// pgpPublicKey returns a PEM encoded PGP public key suitable for testing.
func pgpPublicKey() string {
	return readTestFile("./fastly_test_publickey")
}

// privatekey returns a ASN.1 DER encoded key suitable for testing.
func privateKey() string {
	return readTestFile("./fastly_test_privatekey")
}

// certificate returns a ASN.1 DER encoded certificate for the private key suitable for testing.
func certificate() string {
	return readTestFile("./fastly_test_certificate")
}

// caCert returns a CA certificate suitable for testing
func caCert() string {
	return readTestFile("./fastly_test_cacert")
}

func readTestFile(filename string) string {
	contents, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return string(contents)
}
