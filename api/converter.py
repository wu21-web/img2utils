import os
import subprocess

DEFAULT_BINARY = "bin/img2utils"
DEFAULT_TIMEOUT = 30.0


class ConversionError(RuntimeError):
    pass


def convert(image: bytes, target: str, max_pixels: int = 0) -> bytes:
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
            [binary, "convert", "--to", target, "--max-pixels", str(max_pixels)],
            input=image,
            capture_output=True,
            timeout=timeout,
            check=False,
        )
    except FileNotFoundError as exc:
        raise ConversionError(f"converter binary not found: {binary}") from exc
    except subprocess.TimeoutExpired as exc:
        raise ConversionError("conversion timed out") from exc
    except OSError as exc:
        reason = exc.strerror or str(exc)
        raise ConversionError(f"converter binary is not runnable: {binary} ({reason})") from exc
    except ValueError as exc:
        raise ConversionError(f"invalid conversion request: {exc}") from exc

    if result.returncode != 0:
        detail = result.stderr.decode("utf-8", "replace").strip()
        raise ConversionError(detail or "conversion failed")

    return result.stdout
