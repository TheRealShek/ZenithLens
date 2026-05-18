# ZenithLens — Local Media Gallery

Local-first media gallery: a single Go binary serving an embedded React/Vite frontend.

Quick start

- Dev backend:

  ```bash
  go run main.go
  ```

- Dev frontend (separate terminal):

  ```bash
  cd frontend
  bun install
  bun run dev
  ```

- Production build (frontend -> embed -> binary):

  ```bash
  cd frontend
  bun install
  bun build
  cd ..
  go build -o local-gallery main.go
  ./local-gallery
  ```

Config

- User config lives at `~/.local-gallery/config.json` per project spec.

Repo layout

- `main.go`, `internal/` (backend), `frontend/` (React + Vite).

Where to start

- Read the implementation plan in `Docs/EndGoal.md`.
- Look at `internal/api/handlers.go` for HTTP routes.

Contributing

- Open issues or PRs. Prefer small, focused changes.
