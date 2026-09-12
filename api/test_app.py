import base64
import os
from pathlib import Path

import pytest

import app as app_module

REPO_ROOT = Path(__file__).resolve().parents[1]
SAMPLE_PNG = REPO_ROOT / "internal" / "testdata" / "sample.png"
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
