.PHONY: build clean version

# Determine version from git
VERSION := $(shell git describe --tags --exact-match 2>/dev/null)
ifeq ($(VERSION),)
	GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
	GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
	GIT_STATUS := $(shell git status --untracked-files=no --porcelain 2>/dev/null)
	ifneq ($(GIT_STATUS),)
		VERSION := $(GIT_BRANCH)-$(GIT_COMMIT)-dirty
	else
		VERSION := $(GIT_BRANCH)-$(GIT_COMMIT)
	endif
endif

LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

build:
	@echo "Building bix version $(VERSION)"
	cd cmd/bix && go build $(LDFLAGS) -o bix

build-wasm:
	@echo "Building WASM version $(VERSION)"
	cd cmd/bix/wasm && GOOS=js GOARCH=wasm go build $(LDFLAGS) -o bixscript.wasm

version:
	@echo $(VERSION)

clean:
	rm -f cmd/bix/bix cmd/bix/bix.exe cmd/bix/wasm/bixscript.wasm
