default: fmt lint install generate

# Load .env file if it exists (for VASTAI_API_KEY and other config).
ifneq (,$(wildcard .env))
include .env
export
endif

.PHONY: build install lint fmt test testacc testacc-free sweep generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

fmt:
	gofmt -s -w -e .

test:
	go test -v -count=1 -parallel=4 -timeout 120s ./...

testacc:
	TF_ACC=1 go test -v -count=1 -parallel=4 -timeout 120m -run "TestAcc" ./...

testacc-free:
	TF_ACC=1 go test -v -count=1 -timeout 5m -run "TestAcc" \
		./internal/services/user/... \
		./internal/services/auditlog/... \
		./internal/services/invoice/... \
		./internal/services/offer/... \
		./internal/services/apikey/... \
		./internal/services/envvar/... \
		./internal/services/template/...

sweep:
	@echo "WARNING: This will destroy resources with 'tfacc-' prefix in your Vast.ai account"
	go test ./internal/services/... -v -sweep=all -timeout 15m

generate:
	go generate ./...
