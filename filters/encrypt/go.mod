module github.com/hashicorp/eventlogger/filters/encrypt

go 1.16

require (
	github.com/hashicorp/eventlogger v0.1.0
	github.com/hashicorp/go-kms-wrapping/v2 v2.0.0-20211004181108-59533a548d29
	github.com/hashicorp/go-kms-wrapping/wrappers/aead/v2 v2.0.0-20211004181306-dd5d7c1b481a
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/pointerstructure v1.2.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	google.golang.org/protobuf v1.26.0
)
