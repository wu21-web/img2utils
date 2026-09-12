# img2utils — https://github.com/wu21-web/img2utils
#
# Build from the repository root: docker build -f api/Dockerfile .

FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/img2utils ./cmd/img2utils

FROM python:3.13-slim

LABEL org.opencontainers.image.title="img2utils-flask" \
      org.opencontainers.image.description="Flask HTTP API for the img2utils image converter" \
      org.opencontainers.image.source="https://github.com/wu21-web/img2utils" \
      org.opencontainers.image.url="https://github.com/wu21-web/img2utils" \
      org.opencontainers.image.licenses="MIT"

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    IMG2UTILS_BIN=/app/bin/img2utils \
    HOST=0.0.0.0 \
    PORT=8000

WORKDIR /app

COPY api/requirements.txt api/requirements.txt
RUN pip install --no-cache-dir -r api/requirements.txt

COPY LICENSE LICENSE
COPY api/ api/
COPY --from=builder /out/img2utils bin/img2utils

RUN useradd --create-home --uid 10001 img2utils

USER 10001
EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD ["python", "-c", "import urllib.request; urllib.request.urlopen('http://127.0.0.1:8000/health', timeout=3)"]

CMD ["python", "api/app.py"]
