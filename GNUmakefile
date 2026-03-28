default: fmt lint install generate

.PHONY: build install lint fmt test testacc generate sweep

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

sweep:
	@echo "WARNING: This will destroy resources with 'tfacc-' prefix in your Vast.ai account"
	go test ./internal/services/... -v -sweep=all -timeout 15m

generate:
	go generate ./...
