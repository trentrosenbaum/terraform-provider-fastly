package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

type FastlyTestServiceBlockBackend struct {
	Name    string
	Address string
}


func randomFastlyTestServiceBlockBackend() FastlyTestServiceBlockBackend {
	return FastlyTestServiceBlockBackend{
		Name:    fmt.Sprintf("tf-test-backend-%02d", i),
		Address: fmt.Sprintf("%s.aws.amazon.com", acctest.RandString(3)),
	}
}

func randomFastlyTestServiceBlockBackends(num int) []FastlyTestServiceBlockBackend {
	var backends = []FastlyTestServiceBlockBackend{}
	for i := 0; i < num; i++ {
		backends = append(backends, randomFastlyTestServiceBlockBackend())
	}
	return backends
}