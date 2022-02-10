module github.com/hashicorp/eventlogger/filters/encrypt

go 1.16

require (
	github.com/hashicorp/eventlogger v0.1.0
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.0-20220127162641-13bea7d76bfc
	github.com/hashicorp/go-kms-wrapping/wrappers/aead/v2 v2.0.0-20220210164645-bab1c3a03d9f
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/pointerstructure v1.2.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	google.golang.org/protobuf v1.27.1
)
