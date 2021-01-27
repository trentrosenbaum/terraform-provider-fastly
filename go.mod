module github.com/fastly/terraform-provider-fastly

go 1.14

replace github.com/fastly/go-fastly/v2 v2.1.0 => github.com/opencredo/go-fastly/v2 v2.0.0-20210127124726-6dde4566ce81

require (
	github.com/ajg/form v0.0.0-20160822230020-523a5da1a92f // indirect
	github.com/fastly/go-fastly/v2 v2.1.0
	github.com/google/go-cmp v0.3.0
	github.com/hashicorp/terraform-plugin-sdk v1.1.0
	github.com/stretchr/testify v1.3.0
)
