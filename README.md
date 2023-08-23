# Automation for releasing videos on YouTube

## Run

```bash
go run . --path something.yaml --from-email someone@acme.com \
    --to-thumbnail-email someone@acme.com \
    --to-edit-email someone@acme.com
```

## Build

```bash
go build -o youtube-release

chmod +x youtube-release

sudo mv youtube-release /usr/local/bin
```