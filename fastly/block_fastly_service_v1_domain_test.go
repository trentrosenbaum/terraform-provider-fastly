package fastly

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
)

// Just subset of attributes for now
type FastlyTestServiceBlockDomain struct {
	Name    string
	Address string
}


func randomFastlyTestServiceBlockDomain() FastlyTestServiceBlockDomain {
	return FastlyTestServiceBlockDomain{
		fmt.Sprintf("fastly-test.tf-%s.com", acctest.RandString(10)),
		makeTestServiceComment(),
	}
}
