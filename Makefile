# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean

# Output binary name
BINARY_NAME = youtube-release
# Determine version from git tag or commit hash. Prioritize exact match to a tag.
VERSION := $(shell git describe --tags --exact-match 2>/dev/null || git describe --always --dirty --exclude 'youtube-automation*' 2>/dev/null || echo "dev")

# Cross-compilation settings
OS_ARCHS = \
    darwin/amd64 \
	darwin/arm64 \
    linux/amd64 \
    linux/arm64 \
    windows/amd64

# Build all binaries
all: clean build

# Clean the project
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

# Build binaries for all platforms
build: $(OS_ARCHS)

$(OS_ARCHS):
	$(eval OS := $(word 1, $(subst /, ,$@)))
	$(eval ARCH := $(word 2, $(subst /, ,$@)))
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) -ldflags="-X main.version=$(VERSION)" -o $(BINARY_NAME)_$(OS)_$(ARCH) .

# Build a single binary for the current platform, named simply $(BINARY_NAME)
# This is useful for local testing, like the TestVersionFlag.
build-local:
	CGO_ENABLED=0 $(GOBUILD) -ldflags="-X main.version=$(VERSION)" -o $(BINARY_NAME) .

# Target to bump the patch version and tag
bump-patch:
	$(eval LATEST_TAG := $(shell git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-v:refname | head -n 1))
	$(eval CURRENT_VERSION := $(if $(LATEST_TAG),$(LATEST_TAG),v0.0.0))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'.' '{$$NF = $$NF + 1;} 1' OFS='.'))
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Next version (patch): $(NEXT_VERSION)"
	git tag $(NEXT_VERSION)
	@echo "Tagged with $(NEXT_VERSION)"
	@echo "Run 'git push --tags' to push the new tag."

.PHONY: clean build build-local bump-patch
