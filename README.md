# Automation for releasing videos on YouTube

## Run

```bash
go run . --path something.yaml
```

## Build

```bash
make build # All binaries

go build -o youtube-release # Single binary

chmod +x youtube-release

sudo mv youtube-release /usr/local/bin
```