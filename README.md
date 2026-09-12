# img2utils

Small, dependency-light image conversion commands written in Go.

## Commands

| Command    | Converts     | Status          |
| ---------- | ------------ | --------------- |
| `jpg2png`  | JPEG -> PNG  | Supported       |
| `webp2png` | WebP -> PNG  | Supported       |
| `png2jpg`  | PNG -> JPEG  | Supported       |

WebP decoding is supported; WebP encoding is not.

## Core command

`img2utils convert` reads an image from stdin and writes the converted image to
stdout, which makes it easy to drive from other programs:

```bash
img2utils convert --to png < hello.jpg > hello.png
img2utils convert --to jpg < hello.webp > hello.jpg
```

The target format can be written as `jpg`, `jpeg`, `png`, or `webp`, with or
without a leading dot.

## Usage

```bash
jpg2png hello.jpg -O hello.png
jpg2png hello.jpg -o hello.png
webp2png hello.webp -o hello.png
png2jpg hello.png -O hello.jpg
```

Both `-o` and `-O` are accepted. The output path can be omitted, in which case
the input extension is replaced with the target extension:

```bash
jpg2png hello.jpg   # writes hello.png
webp2png hello.webp # writes hello.png
png2jpg hello.png   # writes hello.jpg
```

## Install

```bash
go install github.com/wu21-web/img2utils/cmd/img2utils@latest
go install github.com/wu21-web/img2utils/cmd/jpg2png@latest
go install github.com/wu21-web/img2utils/cmd/webp2png@latest
go install github.com/wu21-web/img2utils/cmd/png2jpg@latest
```

Requires Go 1.26 or newer.

## Build

```bash
go build ./cmd/img2utils
go build ./cmd/jpg2png
go build ./cmd/webp2png
go build ./cmd/png2jpg
```

## HTTP API

The `api/` directory contains a small Flask service that accepts base64 images
and returns base64 results. Base64 encoding and decoding happen in the HTTP
layer; the Go binary only sees raw bytes over stdin/stdout, so no shell and no
temporary files are involved.

```bash
go build -o bin/img2utils ./cmd/img2utils
python3 -m venv .venv
.venv/bin/pip install -r api/requirements.txt
IMG2UTILS_BIN=bin/img2utils .venv/bin/python api/app.py
```

Send an image by posting its base64 encoding and the target format:

```bash
curl -s http://127.0.0.1:8000/convert \
  -H 'Content-Type: application/json' \
  -d "{\"image\":\"$(openssl base64 -A -in hello.png)\",\"format\":\"jpg\"}"
```

```json
{"format": "jpg", "image": "/9j/4AAQSkZJRgABAQAAAQABAAD..."}
```

`GET /health` returns `{"status": "ok"}`. The request body accepts `image` plus
`format`, `ext`, `?format=`, or `?ext=`; the format may include a leading dot.
Bad requests return `400`, failed conversions return `422`, and bodies larger
than `IMG2UTILS_MAX_UPLOAD` bytes (default 16 MiB) return `413`.

Environment variables:

| Variable               | Default        | Purpose                       |
| ---------------------- | -------------- | ----------------------------- |
| `IMG2UTILS_BIN`        | `bin/img2utils`| Path to the converter binary  |
| `IMG2UTILS_TIMEOUT`    | `30`           | Subprocess timeout in seconds |
| `IMG2UTILS_MAX_UPLOAD` | `16777216`     | Maximum request body in bytes |
| `HOST`                 | `127.0.0.1`    | Dev server bind address       |
| `PORT`                 | `8000`         | Dev server port               |

## Development

```bash
gofmt -l .
go vet ./...
go test ./...
go build -o bin/img2utils ./cmd/img2utils
cd api && ../.venv/bin/pytest
```

## Structure

All conversion logic lives in `internal/convert`. The binaries under `cmd/` are
thin frontends that share the argument parser in `internal/cli`, so new
commands can be added without duplicating conversion or flag handling. The
Flask service in `api/` shells out to `img2utils convert` and owns all base64
handling.

## License

MIT
