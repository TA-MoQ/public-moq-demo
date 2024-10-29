import asyncio
from playwright.async_api import async_playwright, expect
from enum import IntEnum
import os

result_path = "./downloads"
download_timeout = 200000

urls = [
    "https://localhost:1234/?url=https://moq.rorre.me:4443",
    "https://localhost:1234/?url=https://localhost:4443",
    "https://localhost:1234/?url=https://localhost:8443",
    # "https://localhost:1234/?url=https://moq.rorre.me:8443",
]


class StreamMode(IntEnum):
    STREAM = 0
    DATAGRAM = 1
    HYBRID = 2
    AUTO = 3


async def download_results(page, mode, url):
    download_path = os.path.join(result_path, mode)
    print(f"Awaiting for download results to {download_path}...")

    server_url = url.split("?=")[-1]
    server_domain = server_url.strip("https://").split(":")[0]

    # Create downloads directory if it doesn't exist
    os.makedirs(os.path.join(download_path), exist_ok=True)

    downloads = []
    for i in range(4):
        try:
            download = await page.wait_for_event("download", timeout=download_timeout)
            file_path = os.path.join(
                download_path, f"{mode}_{server_domain}_{download.suggested_filename}"
            )
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


async def watch_and_download(url, mode, browser):
    try:
        # Create an isolated browser context and a new page for each task
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
        print(f"Error in watch_and_download for {url} in {mode.name} mode: {e}")
        raise
    finally:
        # Clean up context
        await context.close()


async def main():
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
            # Create tasks for each combination of url and mode
            tasks = [
                watch_and_download(url, mode, browser)
                for url in urls
                for mode in StreamMode
            ]

            # Run all tasks concurrently and wait for them to complete
            await asyncio.gather(*tasks)

        except Exception as e:
            print(f"Error in main execution: {e}")
            raise
        finally:
            # Browser will only close after all tasks (including downloads) are complete
            await browser.close()


# Run the main function
if __name__ == "__main__":
    asyncio.run(main())
