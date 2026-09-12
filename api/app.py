import base64
import binascii
import os

from flask import Flask, jsonify, request

from converter import ConversionError, convert


def env_int(name, default, minimum=None, maximum=None):
    raw = os.environ.get(name)
    if raw is None or not raw.strip():
        return default
    try:
        value = int(raw)
    except ValueError as exc:
        raise RuntimeError(f"{name} must be an integer, got {raw!r}") from exc
    if minimum is not None and value < minimum:
        raise RuntimeError(f"{name} must be at least {minimum}, got {value}")
    if maximum is not None and value > maximum:
        raise RuntimeError(f"{name} must be at most {maximum}, got {value}")
    return value


app = Flask(__name__)
app.config["MAX_CONTENT_LENGTH"] = env_int(
    "IMG2UTILS_MAX_UPLOAD", 16 * 1024 * 1024, minimum=1
)


@app.get("/health")
def health():
    return jsonify(status="ok")


@app.post("/convert")
def convert_image():
    payload = request.get_json(silent=True)
    if not isinstance(payload, dict):
        return jsonify(error="expected a JSON body"), 400

    encoded = payload.get("image")
    if not isinstance(encoded, str) or not encoded.strip():
        return jsonify(error="image must be a base64 string"), 400

    target = requested_format(payload)
    if target is None:
        return jsonify(error="format must be a string"), 400

    try:
        image = base64.b64decode("".join(encoded.split()), validate=True)
    except (binascii.Error, ValueError):
        return jsonify(error="image is not valid base64"), 400

    try:
        converted = convert(image, target)
    except ConversionError as exc:
        return jsonify(error=str(exc)), 422

    return jsonify(
        image=base64.b64encode(converted).decode("ascii"),
        format=target,
    )


def requested_format(payload):
    values = (
        payload.get("format"),
        payload.get("ext"),
        request.args.get("format"),
        request.args.get("ext"),
    )
    for value in values:
        if isinstance(value, str) and value.strip():
            return value.strip().lower()
    return None


if __name__ == "__main__":
    app.run(
        host=os.environ.get("HOST", "127.0.0.1"),
        port=env_int("PORT", 8000, minimum=1, maximum=65535),
    )
