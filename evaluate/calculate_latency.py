from datetime import datetime, timezone
from enum import IntEnum
import io
import json
import os
from pathlib import Path
import re
import shutil
import sys
import tempfile
import time

from playwright.sync_api import sync_playwright, Browser
from PIL import Image
import pytesseract

# 2024-10-14 16:17:09.349
DT_RE = re.compile(r"(\d+-\d+-\d+ \d+:\d+:\d+\.\d+)")


class StreamMode(IntEnum):
    STREAM = 0
    DATAGRAM = 1
    HYBRID = 2
    AUTO = 3


def run_latency_test(browser: Browser, frontend_url: str, mode: StreamMode, path: Path):
    path.mkdir(parents=True, exist_ok=True)

    page = browser.new_page()
    page.goto(frontend_url)
    # Give it some time to load
    time.sleep(5)

    # In a loop because the player SOMETIMES decides to error out and refresh the page
    while page.locator("#start").text_content() == "Start":
        page.click("#start")
        # Give it some time to load
        time.sleep(2)

    # Give it some time to load
    time.sleep(2)
    page.select_option("#category", str(mode.value))
    # Give it some time to load
    time.sleep(2)

    # Ensure we're live
    page.click("#live")

    # Give it some time to load
    time.sleep(5)

    for i in range(4):
        try:
            download = page.wait_for_event("download", timeout=200000)
            file_path = os.path.join(path, download.suggested_filename)
            print(f"Download {i+1} started at {file_path}")
        except Exception as e:
            print(f"Error during download {i+1} in {mode} mode: {e}")
            raise

    results: list[tuple[str, str]] = []
    for _ in range(20):
        page.click("#live")
        time.sleep(2)

        dt = datetime.now(timezone.utc)
        curr = dt.isoformat()
        image_bytes = page.screenshot()
        with open(os.path.join(path, f"{int(dt.timestamp())}.png"), "wb") as f:
            f.write(image_bytes)

        im = Image.open(io.BytesIO(image_bytes))
        extracted: str = pytesseract.image_to_string(im)
        try:
            match = next(DT_RE.finditer(extracted))
            extracted = match.group(1)
        except:
            extracted = "Not found"

        results.append((extracted, curr))
        print(">>>>>>>>>")
        print("Tesseract:", extracted)
        print("Timestamp:", curr)
        time.sleep(10)
    return results


def main(frontend_url: str, output: str):
    with tempfile.TemporaryDirectory() as tmpdirname:
        with sync_playwright() as p:
            browser = p.chromium.launch(
                headless=False,
                args=["--allow-insecure-localhost", "--origin-to-force-quic-on=localhost:4443"],
                executable_path="/usr/bin/google-chrome-stable",
            )
            tmpdir = Path(tmpdirname)
            stream = run_latency_test(browser, frontend_url, StreamMode.STREAM, tmpdir / "stream")
            datagram = run_latency_test(browser, frontend_url, StreamMode.DATAGRAM, tmpdir / "datagram")
            hybrid = run_latency_test(browser, frontend_url, StreamMode.HYBRID, tmpdir / "hybrid")

            browser.close()

        with open(Path(tmpdirname) / "output.json", "w") as f:
            json.dump(
                {
                    "stream": {"latency": stream},
                    "datagram": {"latency": datagram},
                    "hybrid": {"hybrid": hybrid},
                },
                f,
            )

        shutil.make_archive(output, "zip", tmpdirname)


if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(f"Usage: {sys.argv[0]} <frontend> <output>")
        exit()

    main(sys.argv[1], sys.argv[2])
