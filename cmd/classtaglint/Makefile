install:
	@echo "🛠  Building classtaglint"
	@go build -o classtaglint .
	@echo "📦 Moved classtaglint to /usr/local/bin"
	@mv classtaglint /usr/local/bin

test:
	@go test -timeout 30s github.com/hashicorp/eventlogger/cmd/classtaglint/classtagcheck -v