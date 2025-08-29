# Control Panel Web App

A simple configurable web app with a **Go backend** and **plain HTML/CSS** frontend.
The UI is rendered on the server from `config.json` and can include **buttons** and
**sliders**. When activated, the backend calls configured target URLs and reports
success/error back to the UI with temporary toast messages.

---

## Features

* Configurable via JSON (`config.json`)
* Buttons and sliders with labels/icons
* Slider default values fetched server‑side from a backend‑only API
* Backend calls target URLs on activation (button press / slider change)
* Frontend feedback (✅ success / ❌ error toast messages)
* **Live config reload** (backend re-reads `config.json` on every request)
* Dockerfile + Docker Compose for packaging
* Makefile with common commands (build/run/test/etc.)

---

## Project Structure

```
.
├─ main.go
├─ config.json
├─ templates/
│  └─ index.html
├─ static/
│  └─ style.css
├─ Dockerfile
├─ docker-compose.yml
└─ Makefile
```

---

## Prerequisites

* [Go 1.22+](https://golang.org/dl/) (for local development)
* [Docker](https://docs.docker.com/get-docker/)
* [Docker Compose](https://docs.docker.com/compose/)

---

## Quick Start

### Run locally (no Docker)

```bash
make run-local
```

Then open [http://localhost:8080](http://localhost:8080)

### Run with Docker Compose

```bash
make build
make up-d
```

Then open [http://localhost:8080](http://localhost:8080)

Stop everything:

```bash
make down
```

---

## Configuration (`config.json`)

Each UI element is an object with the following properties:

* `name` (string): Display name & identifier
* `type` ("button" | "slider"): Element type
* `icon` (string, optional): Emoji/text icon for buttons
* `min` / `max` (number, sliders only): Range
* `url` (string): Target URL the backend will call on activation

Example:

```json
[
  {
    "name": "Play Music",
    "type": "button",
    "icon": "🎵",
    "url": "http://example.com/play"
  },
  {
    "name": "Pause Music",
    "type": "button",
    "icon": "⏸️",
    "url": "http://example.com/pause"
  },
  {
    "name": "Volume",
    "type": "slider",
    "min": 0,
    "max": 100,
    "url": "http://example.com/volume"
  }
]
```

### Default slider values

On each request, the backend reads `config.json` and fetches the **current default
value per slider** using a backend‑only API (see `fetchDefaultValue` in `main.go`).
Replace the placeholder logic there with your real API call/parse.

---

## Live Reload Behavior

* `config.json` is reloaded on **every HTTP request** (both page render and activation).
* With Docker Compose, `config.json`, `templates/`, and `static/` are **mounted as volumes**
  so edits on the host are reflected immediately inside the container. Just refresh the page.

---

## Frontend UX

* Buttons and sliders are rendered server‑side in a grid:

  * **Buttons**: 2 per row
  * **Sliders**: 1 per row (span 2 columns)
* When you press a button or move a slider, the frontend POSTs to `/activate`.
* Success or error responses display as toast notifications for a few seconds, then fade.
* (Optional) You can throttle/debounce slider events if you want fewer backend calls.

---

## Makefile Commands

Run `make help` to see all commands.

| Command          | Description                                 |
| ---------------- | ------------------------------------------- |
| `make help`      | Show available targets                      |
| `make build`     | Build Docker image(s)                       |
| `make up`        | Run the app in foreground (Docker Compose)  |
| `make up-d`      | Run the app in background (detached)        |
| `make down`      | Stop and remove containers                  |
| `make restart`   | Restart the app                             |
| `make logs`      | Tail container logs                         |
| `make clean`     | Remove containers, images, volumes, orphans |
| `make run-local` | Run Go app locally without Docker           |
| `make test`      | Run Go unit tests with coverage             |

---

## Docker Details

The image is built with a multi‑stage Dockerfile (small final runtime image). Compose mounts
config and static assets read‑only for safety and live editing.

### `Dockerfile`

* Stage 1: Build static Go binary (`CGO_ENABLED=0`)
* Stage 2: Minimal Alpine runtime, copy binary and assets, expose `:8080`

### `docker-compose.yml`

Key parts:

```yaml
services:
  control-panel:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config.json:/app/config.json:ro
      - ./templates:/app/templates:ro
      - ./static:/app/static:ro
    restart: unless-stopped
```

---

## Testing

Run all tests (verbose, with coverage):

```bash
make test
```

---

## Troubleshooting

* **No changes after editing `config.json`**: Ensure it is mounted via Compose and
  refresh the page. The backend re-reads it each request.
* **Network errors on activation**: Toast will show the error text; also check server logs
  (`make logs`). Verify the target URL is reachable from the container/host.
* **Default slider value doesn’t update**: Implement your real API call in `fetchDefaultValue`
  and confirm the API is reachable. Consider timeouts/retries as needed.

---

## License

MIT

