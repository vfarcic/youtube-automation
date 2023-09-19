# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean

# Output binary name
BINARY_NAME = youtube-release

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
	CGO_ENABLED=0 GOOS=$(OS) GOARCH=$(ARCH) $(GOBUILD) -o $(BINARY_NAME)_$(OS)_$(ARCH) .

.PHONY: clean build
