default: fmt lint install generate

.PHONY: build install lint fmt test testacc generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w .

test:
	go test -v -count=1 -parallel=4 -timeout 120s ./...

testacc:
	TF_ACC=1 go test -v -count=1 -parallel=4 -timeout 120m ./...

generate:
	go generate ./...
