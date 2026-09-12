# img2utils

Small, dependency-light image conversion commands written in Go.

## Commands

| Command    | Converts     | Status          |
| ---------- | ------------ | --------------- |
| `jpg2png`  | JPEG -> PNG  | Supported       |
| `webp2png` | WebP -> PNG  | Supported       |
| `png2jpg`  | PNG -> JPEG  | Supported       |

WebP decoding is supported; WebP encoding is not.

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
go install github.com/wu21-web/img2utils/cmd/jpg2png@latest
go install github.com/wu21-web/img2utils/cmd/webp2png@latest
go install github.com/wu21-web/img2utils/cmd/png2jpg@latest
```

Requires Go 1.26 or newer.

## Build

```bash
go build ./cmd/jpg2png
go build ./cmd/webp2png
go build ./cmd/png2jpg
```

## Development

```bash
gofmt -l .
go vet ./...
go test ./...
```

## Structure

All conversion logic lives in `internal/convert`. The binaries under `cmd/` are
thin frontends that share the argument parser in `internal/cli`, so new
commands can be added without duplicating conversion or flag handling.

## License

MIT
