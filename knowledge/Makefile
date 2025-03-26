# get git tag
ifneq ($(GIT_TAG_OVERRIDE),)
$(info GIT_TAG set from env override!)
GIT_TAG := $(GIT_TAG_OVERRIDE)
endif

GIT_TAG   ?= $(shell git describe --tags)
ifeq ($(GIT_TAG),)
GIT_TAG   := $(shell git describe --always)
endif

GO_TAGS := netgo
LD_FLAGS := -s -w -X github.com/obot-platform/tools/knowledge/version.Version=${GIT_TAG}
build:
	go build -mod=mod -o bin/knowledge -tags "${GO_TAGS}" -ldflags '$(LD_FLAGS) ' .

lint:
	golangci-lint run ./...

test:
	go test -v ./...

build-cross:
	GIT_TAG=${GIT_TAG} ./scripts/cross-build.sh

ci-setup:
	@echo "### Installing Go tools..."
	@echo "### -> Installing golangci-lint..."
	curl -sfL $(PKG_GOLANGCI_LINT_SCRIPT) | sh -s -- -b $(GOENVPATH)/bin v$(PKG_GOLANGCI_LINT_VERSION)

gen-checksum:
	$$(cd dist && sha256sum knowledge-* > checksums.txt)