# Bumblebee Maps

Reusable app for generating a "bumblebee map" from Strava / bike data.

The backend is Go/Chi, and the frontend is embedded bun/vue.

## Development Setup

This project uses [mise](https://mise.jdx.co/) to manage development tools and run tasks.

### Prerequisites

Ensure you have `mise` installed. Then, trust the configuration, install the required tools (Go, Bun, Air, DuckDB), and initialize the DuckDB spatial extension:

```bash
mise trust
mise install
mise run setup
```

### Key Commands

Run tasks using `mise run <task>` (or the shorthand `mise r <task>`):

* **Start Development Servers (Backend + Frontend concurrently):**
  ```bash
  mise run dev
  ```
  This runs the backend live-reload server (`air`) and the Vite frontend dev server concurrently.

* **Build for Production:**
  ```bash
  mise run build
  ```
  This builds the Vite/Vue frontend and compiles the Go backend, embedding the frontend assets into the final binary (`bin/bumblebee-bike`).

* **Run Production Build:**
  ```bash
  mise run run
  ```
  Builds and runs the compiled binary.

* **Format Code:**
  ```bash
  mise run fmt
  ```
  Formats all Go and Vue/TS source code in the repository.

* **Clean Build Artifacts:**
  ```bash
  mise run clean
  ```
  Removes the compiled binary and built frontend files.

### Individual Dev Tasks

If you want to run the frontend or backend dev servers separately:
- **Frontend Dev Server:** `mise run dev-frontend` (runs on port 5173, proxies `/api` to port 8080)
- **Backend Dev Server:** `mise run dev-backend` (runs the Go backend using `air` on port 8080)
