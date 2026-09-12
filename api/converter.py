import os
import subprocess

DEFAULT_BINARY = "bin/img2utils"
DEFAULT_TIMEOUT = 30.0


class ConversionError(RuntimeError):
    pass


def convert(image: bytes, target: str) -> bytes:
    binary = os.environ.get("IMG2UTILS_BIN", DEFAULT_BINARY)
    timeout = DEFAULT_TIMEOUT
    raw_timeout = os.environ.get("IMG2UTILS_TIMEOUT")
    if raw_timeout:
        try:
            timeout = float(raw_timeout)
        except ValueError as exc:
            raise ConversionError("IMG2UTILS_TIMEOUT must be a number") from exc

    try:
        result = subprocess.run(
            [binary, "convert", "--to", target],
            input=image,
            capture_output=True,
            timeout=timeout,
            check=False,
        )
    except FileNotFoundError as exc:
        raise ConversionError(f"converter binary not found: {binary}") from exc
    except subprocess.TimeoutExpired as exc:
        raise ConversionError("conversion timed out") from exc

    if result.returncode != 0:
        detail = result.stderr.decode("utf-8", "replace").strip()
        raise ConversionError(detail or "conversion failed")

    return result.stdout
