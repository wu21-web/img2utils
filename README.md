# img2utils

[![CI](https://github.com/wu21-web/img2utils/actions/workflows/ci.yml/badge.svg)](https://github.com/wu21-web/img2utils/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/wu21-web/img2utils/graph/badge.svg?token=CBT6VFMM2K)](https://codecov.io/gh/wu21-web/img2utils)

Small, dependency-light image conversion commands written in Go.

## Commands

| From \ To | JPEG       | PNG        | BMP        |
| --------- | ---------- | ---------- | ---------- |
| WebP      | `webp2jpg` | `webp2png` | `webp2bmp` |
| JPEG      | —          | `jpg2png`  | `jpg2bmp`  |
| PNG       | `png2jpg`  | —          | `png2bmp`  |
| BMP       | `bmp2jpg`  | `bmp2png`  | —          |

The input format is detected from the file contents, so any command reads any
supported input; the name describes the common case. WebP is decode-only, which
is why nothing writes WebP.

## Core command

`img2utils convert` reads an image from stdin and writes the converted image to
stdout, which makes it easy to drive from other programs:

```bash
img2utils convert --to png < hello.jpg > hello.png
img2utils convert --to jpg < hello.webp > hello.jpg
img2utils convert --to bmp < hello.png > hello.bmp
```

The target format can be written as `jpg`, `jpeg`, `png`, `bmp`, or `webp`, with
or without a leading dot.

## Formats

The input format is detected from the image contents, so any supported input can
be converted to any supported output:

| Format | Decode | Encode |
| ------ | ------ | ------ |
| JPEG   | yes    | yes    |
| PNG    | yes    | yes    |
| BMP    | yes    | yes    |
| WebP   | yes    | no     |

## Usage

```bash
jpg2png hello.jpg -O hello.png
jpg2png hello.jpg -o hello.png
webp2png hello.webp -o hello.png
png2jpg hello.png -O hello.jpg
png2bmp hello.png -O hello.bmp
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
go install github.com/wu21-web/img2utils/cmd/bmp2jpg@latest
go install github.com/wu21-web/img2utils/cmd/bmp2png@latest
go install github.com/wu21-web/img2utils/cmd/jpg2bmp@latest
go install github.com/wu21-web/img2utils/cmd/jpg2png@latest
go install github.com/wu21-web/img2utils/cmd/png2bmp@latest
go install github.com/wu21-web/img2utils/cmd/png2jpg@latest
go install github.com/wu21-web/img2utils/cmd/webp2bmp@latest
go install github.com/wu21-web/img2utils/cmd/webp2jpg@latest
go install github.com/wu21-web/img2utils/cmd/webp2png@latest
```

Requires Go 1.26 or newer.

## Build

```bash
go build ./cmd/...
```

## HTTP API

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
{ "format": "jpg", "image": "/9j/4AAQSkZJRgABAQAAAQABAAD..." }
```

`GET /health` returns `{"status": "ok"}`. The request body accepts `image` plus
`format`, `ext`, `?format=`, or `?ext=`; the format may include a leading dot.
Bad requests return `400`, failed conversions return `422`, and bodies larger
than `IMG2UTILS_MAX_UPLOAD` bytes (default 16 MiB) return `413`.

Environment variables:

| Variable               | Default         | Purpose                       |
| ---------------------- | --------------- | ----------------------------- |
| `IMG2UTILS_BIN`        | `bin/img2utils` | Path to the converter binary  |
| `IMG2UTILS_TIMEOUT`    | `30`            | Subprocess timeout in seconds |
| `IMG2UTILS_MAX_UPLOAD` | `16777216`      | Maximum request body in bytes |
| `IMG2UTILS_MAX_PIXELS` | `25000000`      | Decoded image pixel limit     |
| `HOST`                 | `127.0.0.1`     | Dev server bind address       |
| `PORT`                 | `8000`          | Dev server port               |

Malformed integer values for these variables stop startup with a clear error
instead of a bare traceback. Images whose declared dimensions exceed
`IMG2UTILS_MAX_PIXELS` are rejected before decoding, which keeps a small
compressed file from being expanded into a huge pixel buffer.

### Docker

```bash
docker build -f api/Dockerfile -t img2utils-flask .
docker run --rm -p 8000:8000 img2utils-flask
```

```bash
docker run --rm -p 8000:8000 ghcr.io/wu21-web/img2utils-flask:v0.0
```

Every push to `main` also builds the image and moves the `latest` tag, so that
follows the tip of `main` independently of releases:

```bash
docker run --rm -p 8000:8000 ghcr.io/wu21-web/img2utils-flask:latest
```

## Development

```bash
gofmt -l .
go vet ./...
go test ./...
go build -o bin/img2utils ./cmd/img2utils
python3 -m venv .venv
.venv/bin/pip install -r api/requirements-dev.txt
cd api && ../.venv/bin/pytest
```

## License

MIT
