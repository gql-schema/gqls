# Make
.DEFAULT_GOAL  := help

# Versioning
VERSION        ?= $(shell git describe --tags --dirty --always 2>/dev/null || echo "dev")
COMMIT         := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")

# Go
GO             ?= go
PKGS           := $(shell $(GO) list ./... | grep -v /vendor/)
BINARY         ?= gqls
ENTRYPOINT     ?= ./cmd/cli/main.go
WASM_BINARY    := ./web/main.wasm
WASM_ENTRYPOINT := ./cmd/wasm/main.go

# Tools
TOOLS_DIR      := tools
TOOLS_GO       := $(TOOLS_DIR)/go.mod
ANTLR_VERSION  ?= 4.13.2
ANTLR_JAR      := $(TOOLS_DIR)/antlr-$(ANTLR_VERSION)-complete.jar

# Help function taken from: https://gist.githubusercontent.com/prwhite/8168133/raw/c3055b98fc9e51f74d082b82c0b571fbbd7a7b0e/Makefile
help:
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

.PHONY: tidy
tidy:  	                    ## Install dependencies
	$(GO) mod tidy

.PHONY: tools-tidy
tools-tidy:
	cd $(TOOLS_DIR); $(GO) mod tidy

.PHONY: fmt
fmt:                        ## Format code
fmt: tools-tidy
	$(GO) tool -modfile=$(TOOLS_GO) gofumpt -l -w .

.PHONY: lint
lint:                       ## Run linters
lint: tools-tidy
	$(GO) tool -modfile=$(TOOLS_GO) golangci-lint run ./...
	$(GO) tool -modfile=$(TOOLS_GO) go-check-sumtype ./...

$(ANTLR_JAR):
	mkdir -p $(dir $@)
	curl -L "https://www.antlr.org/download/antlr-$(ANTLR_VERSION)-complete.jar" -o $@

.PHONY: generate
generate:                   ## Generate code
generate: $(ANTLR_JAR)
	mkdir -p internal/gql/grammar
	cd $(TOOLS_DIR); $(GO) generate ./...

.PHONY: build
build:                      ## Build the binary
build: generate
	CGO_ENABLED=0 $(GO) build -o $(BINARY) -ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" $(ENTRYPOINT)

.PHONY: build-static-site
build-static-site: generate ## Build the static site
	cp examples/schema.gql web/example-schema.gql
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" web/wasm_exec.js
	GOOS=js GOARCH=wasm $(GO) build -ldflags "-s -w -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)'" -o $(WASM_BINARY) $(WASM_ENTRYPOINT)

.PHONY: serve
serve: build-static-site    ## Serve the web UI
	cd web; python3 -m http.server 8000

.PHONY: test
test:                       ## Run unit tests
test: generate
	$(GO) test -race -covermode=atomic -coverprofile=coverage.out ./... -v

.PHONY: test-integration
test-integration:           ## Run unit and integration tests
test-integration: generate
	$(GO) test -race -tags=integration ./... -v

.PHONY: clean
clean:                      ## Clean up build artifacts
	rm -rf $(BINARY) coverage.out $(ANTLR_JAR) internal/web/wasm/ docs/main.wasm docs/wasm_exec.js web/example-schema.gql
