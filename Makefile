test: 
	go test ./... -timeout 30m

fmt:
	gofumpt -w $$(find . -name '*.go')

tools:
	go generate -tags tools tools/tools.go