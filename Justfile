timeout := "300s"

# List tasks.
default:
  just --list

# Build the binary and move it to `/usr/local/bin`.
build:
  go build -o youtube-release ./cmd/youtube-automation
  chmod +x youtube-release
  sudo mv youtube-release /usr/local/bin

# Build a local binary for testing (without installing to system)
build-local:
  go build -o youtube-release ./cmd/youtube-automation
  chmod +x youtube-release

# Run the application with environment variables
run:
  eval $(vals env -export -f .env.vals.yaml)
  youtube-release

# Run tests with coverage
test:
  go test ./... -cover

# Clean build artifacts
clean:
  rm -f youtube-release
