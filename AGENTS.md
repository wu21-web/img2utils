# Repository Guidelines

## Project Structure & Module Organization

- `cmd/` — one package per command: nine conversions (`jpg2png`, `bmp2png`, `webp2jpg`, ...) plus the `img2utils` umbrella.
- `internal/cli/` — shared flag parsing, the command table in `commands.go`, and its tests.
- `internal/convert/` — format parsers and encoders (`jpeg.go`, `png.go`, `bmp.go`, `webp.go`); all conversion logic lives here.
- `internal/testdata/` — fixtures such as `sample.png`, `sample.webp`, `oversized.png`.
- `api/` — Flask service (`app.py`, `converter.py`), pytest suite, and `Dockerfile`.
- `.github/workflows/` — `ci.yml` (test, build, api, container) and `release.yml`.

## Build, Test, and Development Commands

```bash
gofmt -l .                # formatting check
go vet ./...              # static analysis
go test ./...             # unit tests
go test -covermode=atomic -coverprofile=coverage.txt ./...   # coverage
go build ./cmd/...        # build every binary
go run ./cmd/bmp2png photo.webp        # writes photo.png
```

```bash
go build -trimpath -o bin/img2utils ./cmd/img2utils
python3 -m venv .venv && .venv/bin/pip install -r api/requirements-dev.txt
cd api && ../.venv/bin/pytest          # API suite (needs the binary above)
docker build -f api/Dockerfile -t img2utils-flask .   # context is the repo root
```

## Coding Style & Naming Conventions

- Use `gofmt` output, tabs, and standard Go naming; comment only exported identifiers.
- Prefer the standard library, adding `golang.org/x/image` as the sole image dependency. Parse flags with `flag`, not a framework.
- Wrap errors with `%w`, compare with `errors.Is`.
- Keep command packages thin: `cmd/<name>/main.go` only wires up `internal/cli`.

## Testing Guidelines

- Go tests use the `testing` package; API tests use `pytest` with `pytest-cov`.
- Name tests `TestXxx`, prefer table-driven cases, and write file output into `t.TempDir()`.
- Adding a command means an entry in `internal/cli/commands.go` plus `cmd/<name>/main.go`; `TestCommandsMatchCmdPackages` fails when those disagree.
- Hold patch coverage at the 86% Codecov target; cover error and collision paths too.

## Commit & Pull Request Guidelines

- Subjects are short and imperative, optionally prefixed: `feat(cli): support glob patterns`, `chore(ci): ...`, `Add BMP and remaining conversion commands`.
- No co-author trailers, generated files, or IDE settings in commits.
- Open PRs against `main` after `gofmt`, `go vet`, `go test`, and the `api/` suite pass.
- Describe user-visible changes (flags, exit codes, formats) with examples.

## Runtime Notes

- Formats: JPEG, PNG, and BMP encode and decode; WebP is decode-only (`x/image/webp` has no encoder). GIF, TIFF, and AVIF are unsupported.
- Secrets: `CODECOV_TOKEN` for coverage uploads, `DOCKER_TOKEN` for GHCR pushes.
- Pushes to `main` publish `ghcr.io/wu21-web/img2utils-flask:latest`; version tags publish the tag plus semver forms.
