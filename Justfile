timeout := "300s"

# List tasks.
default:
  just --list

# Build the binary and move it to `/usr/local/bin`.
build: frontend-build
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

# Build the frontend (npm install + build) and copy to Go embed directory
frontend-build:
  cd web && npm install && npm run build
  rm -rf internal/frontend/dist
  cp -r web/dist internal/frontend/dist

# Build the full binary with embedded frontend
build-full: frontend-build build-local

# Clean build artifacts
clean:
  rm -f youtube-release
