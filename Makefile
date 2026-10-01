.SILENT:

run:
	air

update:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run ./cmd/... ./internal/...

test:
	go test ./cmd/... ./internal/...
