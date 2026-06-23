# send2ereader

A lightweight Go backend for transferring files from a computer to Kobo or Kindle e-readers using a short temporary code. The app uses Go's embedded HTML templates and HTMX for a single-language frontend/backend experience.

## Features

- Detects Kobo and Kindle user agents
- Generates a temporary 5-character code for e-readers
- Upload form for desktop browsers
- HTMX-powered polling for e-reader status updates
- In-memory ephemeral transfer sessions
- No external database required

## Run

```bash
cd /Users/austinarcher/Code/Go/send2ereader
go mod tidy
go run .
```

Then open `http://localhost:8080`.

## Notes

- Uploaded files are saved to the OS temp directory and cleaned up automatically after inactivity.
- The conversion target is tracked, but the current implementation preserves the uploaded file and makes it available for download.
