import asyncio
from playwright.async_api import async_playwright, expect
from enum import IntEnum
import os
import time

result_path = "./downloads"
download_timeout = 200000

urls = [
    # "https://localhost:1234/?url=https://moq.rorre.me:4443",
    "https://localhost:1234/?url=https://localhost:4443",
    "https://localhost:1234/?url=https://localhost:8443",
    # "https://localhost:1234/?url=https://moq.rorre.me:8443",
]


class StreamMode(IntEnum):
    STREAM = 0
    DATAGRAM = 1
    HYBRID = 2
    # AUTO = 3


async def download_results(page, mode, url):
    server_url = url.split("?url=")[-1]
    server_domain = server_url.strip("https://").replace(":", "_").replace(".", "_")
    download_path = os.path.join(result_path, mode, server_domain)

    print(
        f"Awaiting for download results to {download_path} for {mode} in {server_domain}..."
    )

    # Create downloads directory if it doesn't exist
    os.makedirs(os.path.join(download_path), exist_ok=True)

    downloads = []
    for i in range(5):
        try:
            download = await page.wait_for_event("download", timeout=download_timeout)
            file_path = os.path.join(download_path, download.suggested_filename)
            # Store the download promise instead of awaiting it immediately
            downloads.append(download.save_as(file_path))
            print(f"Download {i+1} started at {file_path}")
        except Exception as e:
            print(f"Error during download {i+1} at {url} in {mode} mode: {e}")
            raise

    # Wait for all downloads to complete
    try:
        await asyncio.gather(*downloads)
        print(f"All downloads completed for {download_path}")
    except Exception as e:
        print(f"Error while waiting for downloads to complete: {e}")
        raise


async def process_url(url, browser):
    print(f"\n=== Processing URL: {url} ===\n")

    for mode in StreamMode:
        try:
            # Create an isolated browser context and a new page for each mode
            context = await browser.new_context(ignore_https_errors=True)
            page = await context.new_page()

            await page.goto(url)
            await asyncio.sleep(5)  # Allow time for the page to load

            while await page.locator("#start").text_content() == "Start":
                await page.click("#start")
                await asyncio.sleep(2)  # Allow time for playback to start

            print(f"[{mode.name} - {url}] Playback started...")
            await asyncio.sleep(2)  # Additional wait for stable playback
            await page.select_option("#category", str(mode.value))

            await expect(page.locator("label.audio.label span")).not_to_have_text("")
            await expect(page.locator("label.video.label span")).not_to_have_text("")
            await expect(page.locator("#start")).to_have_text("Stop")

            print(f"[{mode.name} - {url}] All checks passed!")

            await download_results(page, f"{mode.name.lower()}", url)

            await page.mouse.wheel(0, 500)
            await page.evaluate("document.body.style.zoom=0.65")

        except Exception as e:
            print(f"Error processing {url} in {mode.name} mode: {e}")
            raise
        finally:
            await context.close()
            print(f"\n=== Completed {mode.name} mode for {url} ===\n")


async def main():
    start_time = time.time()
    async with async_playwright() as playwright:
        browser = await playwright.chromium.launch(
            args=[
                "--allow-insecure-localhost",
                "--ignore-certificate-errors-spki-list=tsZq5OptbVTW7lgmwGFVR7IBUyVE1x6R3ZBUujlEkOc=",
                "--origin-to-force-quic-on=localhost:4443",
            ],
            executable_path="C:\Program Files\Google\Chrome\Application\chrome.exe",
        )

        try:
            # Process one URL at a time
            for url in urls:
                try:
                    await process_url(url, browser)
                except Exception as e:
                    print(f"Failed to process URL {url}: {e}")
                    continue  # Continue with next URL even if one fails

                print(f"\n=== Completed all modes for URL: {url} ===\n")

        except Exception as e:
            print(f"Error in main execution: {e}")
            raise
        finally:
            await browser.close()
            print(f"Total execution time: {time.time() - start_time} seconds")


if __name__ == "__main__":
    asyncio.run(main())
