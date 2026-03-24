default: build

BINARY        = terraform-provider-openrouter
NAMESPACE     = braveness23
PROVIDER_NAME = openrouter
VERSION       = dev
OS_ARCH       = $(shell go env GOOS)_$(shell go env GOARCH)
INSTALL_PATH  = ~/.terraform.d/plugins/registry.terraform.io/$(NAMESPACE)/$(PROVIDER_NAME)/$(VERSION)/$(OS_ARCH)

.PHONY: build
build:
	go build -o $(BINARY) .

.PHONY: install
install: build
	mkdir -p $(INSTALL_PATH)
	mv $(BINARY) $(INSTALL_PATH)/$(BINARY)

.PHONY: test
test:
	go test ./... -v -count=1 -timeout 30s

.PHONY: testacc
testacc:
	TF_ACC=1 go test ./internal/provider/... -v -count=1 -timeout 120m

.PHONY: fmt
fmt:
	gofmt -s -w .
	goimports -w .

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: generate
generate:
	go generate ./...

.PHONY: docs
docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name openrouter

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -f $(BINARY)
