module github.com/hashicorp/eventlogger

go 1.16

require (
	github.com/IBM/sarama v1.43.2
	github.com/go-test/deep v1.0.4
	github.com/hashicorp/go-multierror v1.1.1
	github.com/hashicorp/go-secure-stdlib/base62 v0.1.1
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.1
	github.com/hashicorp/go-uuid v1.0.3
	github.com/stretchr/testify v1.9.0
	github.com/testcontainers/testcontainers-go v0.30.0
	github.com/testcontainers/testcontainers-go/modules/kafka v0.30.0
	go.uber.org/goleak v1.2.1
	mvdan.cc/gofumpt v0.1.1
)
