timeout := "300s"

# List tasks.
default:
  just --list

# Build the binary and move it to `/usr/local/bin`.
build:
  go build -o youtube-release
  chmod +x youtube-release
  sudo mv youtube-release /usr/local/bin

run:
  eval "$(teller sh)"
  youtube-release
