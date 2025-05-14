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
	rm -f $(BINARY_NAME) $(BINARY_NAME)_*.*

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

# Common logic to get current version or default
GET_CURRENT_VERSION = $(strip $(shell git tag --list "v[0-9]*.[0-9]*.[0-9]*" --sort=-v:refname | head -n 1 || echo "v0.0.0"))

# Macro to perform the tagging and echo messages
define BUMP_AND_TAG
    @echo "Current version: $(1)"
    @echo "Next version ($(2) bump): $(3)"
    git tag $(3)
    @echo "Tagged with $(3)"
    @echo "Run 'git push --tags' to push the new tag."
endef

# Target to echo the next minor version string without tagging
get-next-minor-version:
	$(eval CURRENT_VERSION := $(call GET_CURRENT_VERSION))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'[v.]' '{printf "v%d.%d.0", $$2, $$3+1}'))
	@echo $(NEXT_VERSION)

# Target to echo the next patch version string without tagging
get-next-patch-version:
	$(eval CURRENT_VERSION := $(call GET_CURRENT_VERSION))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'[v.]' '{printf "v%d.%d.%d", $$2, $$3, $$4+1}'))
	@echo $(NEXT_VERSION)

# Target to echo the next major version string without tagging
get-next-major-version:
	$(eval CURRENT_VERSION := $(call GET_CURRENT_VERSION))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'[v.]' '{printf "v%d.0.0", $$2+1}'))
	@echo $(NEXT_VERSION)

# Default: Bump minor version
bump-version:
	$(eval CURRENT_VERSION := $(call GET_CURRENT_VERSION))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'[v.]' '{printf "v%d.%d.0", $$2, $$3+1}'))
	$(call BUMP_AND_TAG,$(CURRENT_VERSION),minor,$(NEXT_VERSION))

bump-patch:
	$(eval CURRENT_VERSION := $(call GET_CURRENT_VERSION))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'[v.]' '{printf "v%d.%d.%d", $$2, $$3, $$4+1}'))
	$(call BUMP_AND_TAG,$(CURRENT_VERSION),patch,$(NEXT_VERSION))

bump-major:
	$(eval CURRENT_VERSION := $(call GET_CURRENT_VERSION))
	$(eval NEXT_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F'[v.]' '{printf "v%d.0.0", $$2+1}'))
	$(call BUMP_AND_TAG,$(CURRENT_VERSION),major,$(NEXT_VERSION))

.PHONY: clean build build-local bump-version bump-patch bump-major get-next-minor-version get-next-patch-version get-next-major-version
