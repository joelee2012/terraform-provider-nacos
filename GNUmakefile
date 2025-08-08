default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

testacc:
	TF_ACC=1 go test -v -coverprofile=coverage.txt -covermode=atomic -timeout 120m ./...
	go tool cover -html=coverage.txt -o cover.html

.PHONY: fmt lint test testacc build install generate
