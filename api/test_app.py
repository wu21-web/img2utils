import base64
import os
import subprocess
import sys
from pathlib import Path

import pytest

import app as app_module

REPO_ROOT = Path(__file__).resolve().parents[1]
API_ROOT = Path(__file__).resolve().parent
SAMPLE_PNG = REPO_ROOT / "internal" / "testdata" / "sample.png"
OVERSIZED_PNG = REPO_ROOT / "internal" / "testdata" / "oversized.png"
BINARY = Path(os.environ.get("IMG2UTILS_BIN", REPO_ROOT / "bin" / "img2utils"))


@pytest.fixture(scope="session", autouse=True)
def converter_binary():
    if not BINARY.exists():
        pytest.skip(f"converter binary not built: {BINARY}")
    os.environ["IMG2UTILS_BIN"] = str(BINARY)


@pytest.fixture
def client():
    app_module.app.config.update(TESTING=True)
    return app_module.app.test_client()


def sample_base64():
    return base64.b64encode(SAMPLE_PNG.read_bytes()).decode("ascii")


def test_health(client):
    response = client.get("/health")

    assert response.status_code == 200
    assert response.get_json() == {"status": "ok"}


def test_convert_png_to_jpg(client):
    response = client.post(
        "/convert",
        json={"image": sample_base64(), "format": "jpg"},
    )

    assert response.status_code == 200
    body = response.get_json()
    assert body["format"] == "jpg"
    assert base64.b64decode(body["image"]).startswith(b"\xff\xd8")


def test_convert_accepts_ext_field(client):
    response = client.post(
        "/convert",
        json={"image": sample_base64(), "ext": ".png"},
    )

    assert response.status_code == 200
    assert base64.b64decode(response.get_json()["image"]).startswith(b"\x89PNG")


def test_convert_accepts_query_format(client):
    response = client.post(
        "/convert?format=jpg",
        json={"image": sample_base64()},
    )

    assert response.status_code == 200
    assert base64.b64decode(response.get_json()["image"]).startswith(b"\xff\xd8")


def test_convert_rejects_non_json_body(client):
    response = client.post("/convert", data="not json")

    assert response.status_code == 400
    assert "error" in response.get_json()


def test_convert_rejects_missing_image(client):
    response = client.post("/convert", json={"format": "png"})

    assert response.status_code == 400
    assert "error" in response.get_json()


def test_convert_rejects_missing_format(client):
    response = client.post("/convert", json={"image": sample_base64()})

    assert response.status_code == 400
    assert "error" in response.get_json()


def test_convert_rejects_invalid_base64(client):
    response = client.post(
        "/convert",
        json={"image": "not base64!", "format": "png"},
    )

    assert response.status_code == 400
    assert "error" in response.get_json()


def test_convert_rejects_unknown_format(client):
    response = client.post(
        "/convert",
        json={"image": sample_base64(), "format": "bmp"},
    )

    assert response.status_code == 422
    assert "unsupported format" in response.get_json()["error"]


def test_convert_reports_webp_encoding_error(client):
    response = client.post(
        "/convert",
        json={"image": sample_base64(), "format": "webp"},
    )

    assert response.status_code == 422
    assert "webp encoding" in response.get_json()["error"]


def test_convert_reports_invalid_image(client):
    response = client.post(
        "/convert",
        json={
            "image": base64.b64encode(b"not an image").decode("ascii"),
            "format": "png",
        },
    )

    assert response.status_code == 422
    assert "decode" in response.get_json()["error"]


def test_convert_reports_missing_binary(client, tmp_path, monkeypatch):
    monkeypatch.setenv("IMG2UTILS_BIN", str(tmp_path / "missing"))

    response = client.post("/convert", json={"image": sample_base64(), "format": "png"})

    assert response.status_code == 422
    assert "not found" in response.get_json()["error"]


def test_convert_reports_non_executable_binary(client, tmp_path, monkeypatch):
    binary = tmp_path / "img2utils"
    binary.write_bytes(b"#!/bin/sh\n")
    binary.chmod(0o644)
    monkeypatch.setenv("IMG2UTILS_BIN", str(binary))

    response = client.post("/convert", json={"image": sample_base64(), "format": "png"})

    assert response.status_code == 422
    assert "not runnable" in response.get_json()["error"]


def test_convert_reports_binary_directory(client, tmp_path, monkeypatch):
    monkeypatch.setenv("IMG2UTILS_BIN", str(tmp_path))

    response = client.post("/convert", json={"image": sample_base64(), "format": "png"})

    assert response.status_code == 422
    assert "not runnable" in response.get_json()["error"]


def test_convert_rejects_oversized_image(client):
    bomb = base64.b64encode(OVERSIZED_PNG.read_bytes()).decode("ascii")

    response = client.post("/convert", json={"image": bomb, "format": "png"})

    assert response.status_code == 422
    assert "too many pixels" in response.get_json()["error"]


def test_convert_enforces_configured_pixel_limit(client, monkeypatch):
    monkeypatch.setitem(app_module.app.config, "IMG2UTILS_MAX_PIXELS", 1)

    response = client.post("/convert", json={"image": sample_base64(), "format": "png"})

    assert response.status_code == 422
    assert "too many pixels" in response.get_json()["error"]


def test_convert_rejects_nul_in_format(client):
    response = client.post(
        "/convert",
        json={"image": sample_base64(), "format": "png\u0000"},
    )

    assert response.status_code == 422
    assert "invalid conversion request" in response.get_json()["error"]


def test_env_int_uses_default(monkeypatch):
    monkeypatch.delenv("PORT", raising=False)

    assert app_module.env_int("PORT", 8000) == 8000


def test_env_int_reads_value(monkeypatch):
    monkeypatch.setenv("PORT", "9000")

    assert app_module.env_int("PORT", 8000) == 9000


def test_env_int_rejects_non_integer(monkeypatch):
    monkeypatch.setenv("PORT", "abc")

    with pytest.raises(RuntimeError, match="PORT must be an integer"):
        app_module.env_int("PORT", 8000)


def test_env_int_rejects_out_of_range(monkeypatch):
    monkeypatch.setenv("PORT", "70000")

    with pytest.raises(RuntimeError, match="at most 65535"):
        app_module.env_int("PORT", 8000, minimum=1, maximum=65535)


def test_invalid_max_upload_fails_with_clear_message():
    result = subprocess.run(
        [sys.executable, "-c", "import app"],
        cwd=API_ROOT,
        env={**os.environ, "IMG2UTILS_MAX_UPLOAD": "abc"},
        capture_output=True,
        text=True,
    )

    assert result.returncode != 0
    assert "IMG2UTILS_MAX_UPLOAD must be an integer" in result.stderr


def test_invalid_max_pixels_fails_with_clear_message():
    result = subprocess.run(
        [sys.executable, "-c", "import app"],
        cwd=API_ROOT,
        env={**os.environ, "IMG2UTILS_MAX_PIXELS": "abc"},
        capture_output=True,
        text=True,
    )

    assert result.returncode != 0
    assert "IMG2UTILS_MAX_PIXELS must be an integer" in result.stderr
