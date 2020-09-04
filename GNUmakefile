SHELL = bash
PROJECT_ROOT := $(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
THIS_OS := $(shell uname)

BUILDDIR=pkg
PKG_NAME=keylightctl

GIT_COMMIT := $(shell git rev-parse HEAD)
GIT_DIRTY := $(if $(shell git status --porcelain),+CHANGES)
GO_LDFLAGS := "-X github.com/endocrimes/keylightctl/version.GitCommit=$(GIT_COMMIT)$(GIT_DIRTY)"

GO := go

GOOSARCHES += linux/386 \
	linux/amd64 \
	linux/arm \
	linux/arm64 \
	windows/386 \
	windows/amd64 \
	darwin/amd64

default: help

define buildtarget
$(BUILDDIR)/$(1)_$(2)/$(PKG_NAME):
	@echo "==> Building $(BUILDDIR)/$(1)_$(2)/$(PKG_NAME)..."
	@mkdir -p $(BUILDDIR)/$(1)_$(2)
	@GOOS=$(1) GOARCH=$(2) $(GO) build \
			 -o $(BUILDDIR)/$(1)_$(2)/$(PKG_NAME) \
			 -a -ldflags $(GO_LDFLAGS) -tags "$(GO_TAGS)";
.PHONY: $(BUILDDIR)/$(1)_$(2)/$(PKG_NAME)
endef

$(foreach GOOSARCH,$(GOOSARCHES),$(eval $(call buildtarget,$(subst /,,$(dir $(GOOSARCH))),$(notdir $(GOOSARCH)))))

.PHONY: all
all: $(foreach GOOSARCH,$(GOOSARCHES),$(BUILDDIR)/$(subst /,,$(dir $(GOOSARCH)))_$(notdir $(GOOSARCH))/$(PKG_NAME)) ## Build for all supported platforms

.PHONY: fmt
fmt: ## Run gofmt over all source files
	@echo "==> Fixing source files with gofmt..."
	@find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

.PHONY: build
build: GOOS=$(shell go env GOOS)
build: GOARCH=$(shell go env GOARCH)
build: GOPATH=$(shell go env GOPATH)
build: DEV_TARGET=$(BUILDDIR)/$(GOOS)_$(GOARCH)/$(PKG_NAME)
build: fmt ## Build for the current development platform
	@echo "==> Removing old development build"
	@rm -f $(PROJECT_ROOT)/$(DEV_TARGET)
	@rm -f $(PROJECT_ROOT)/bin/$(PKG_NAME)
	@rm -f $(GOPATH)/bin/$(PKG_NAME)
	@$(MAKE) --no-print-directory \
		$(DEV_TARGET)
	@mkdir -p $(PROJECT_ROOT)/bin
	@mkdir -p $(GOPATH)/bin
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(PROJECT_ROOT)/bin/
	@cp $(PROJECT_ROOT)/$(DEV_TARGET) $(GOPATH)/bin/


.PHONY: tools
tools: GOBIN=$(PROJECT_ROOT)/tools/bin
tools:
	@echo "==> Installing tools to tools/bin"
	@mkdir -p $(GOBIN)
	@GOBIN=$(GOBIN) go env
	@cd tools && GOBIN=$(GOBIN) go install github.com/goreleaser/goreleaser

HELP_FORMAT="    \033[36m%-25s\033[0m %s\n"
.PHONY: help
help: ## Display this usage information
	@echo "Valid targets:"
	@grep -E '^[^ ]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; \
			{printf $(HELP_FORMAT), $$1, $$2}'
	@echo ""
