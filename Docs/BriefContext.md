# ZenithLens - Brief Context

## What Is This
Local-only media gallery. Single Go binary serves a React frontend. Users register filesystem folders, the app scans them and presents a browsable grid with thumbnails, lightbox, search, favorites, and pagination. No cloud, no auth, no database.

## Tech Stack
- **Backend:** Go (stdlib net/http), single binary
- **Frontend:** React + Vite, built to static assets, embedded into Go binary via `//go:embed frontend/dist`
- **Package Manager:** Bun (npm works as fallback)
- **Thumbnails:** `disintegration/imaging` (images), `ffmpeg` CLI (videos, optional)
- **Config:** JSON at `~/.local-gallery/config.json`
- **Thumb Cache:** Disk at `~/.local-gallery/thumbs/`

## Backend Structure
| Package | File | Responsibility |
|---------|------|----------------|
| `internal/media` | `types.go` | Extensions, MIME, MediaFile struct, IsMedia |
| `internal/config` | `config.go` | Load/Save config (atomic), folder/favorite helpers |
| `internal/scanner` | `scanner.go` | Context-aware WalkDir, symlink/hidden skip |
| `internal/api` | `dto.go` | FolderDTO, MediaFileDTO, PageResponseDTO |
| `internal/api` | `handlers.go` | All endpoints, path validation, scan lifecycle, routing |
| `internal/thumb` | `thumb.go` | singleflight, imaging resize, ffmpeg timeout, temp cleanup |
| root | `main.go` | Embed, flags (--open), signal handling, startup scans |

## Frontend Structure
| Layer | Files | Responsibility |
|-------|-------|----------------|
| Entry | `main.jsx` | React root mount |
| Layout | `App.jsx`, `App.css` | Thin orchestrator, global styles/variables |
| Context | `context/AppContext.jsx` | Shared state: folders, favorites, scanning, polling |
| Hooks | `hooks/useRouteState.js` | URL-synced route/page/search state |
| Hooks | `hooks/useMediaFetch.js` | Data fetching for current view |
| Hooks | `hooks/useLightbox.js` | Lightbox state + 5 keyboard shortcuts |
| Components | `Sidebar, Toolbar, Grid, MediaItem, Pagination, Lightbox, AddFolderModal` | Each with co-located .css |

## API Endpoints
- `GET/POST /api/folders` — list/add folders
- `DELETE /api/folders/:id` — remove folder
- `POST /api/folders/:id/rescan` — trigger rescan
- `GET /api/home` — all media, paginated, shuffled
- `GET /api/folder/:id` — single folder media
- `GET /api/favorites`, `POST /api/favorites`, `DELETE /api/favorites?path=` — favorites CRUD
- `GET /api/search?q=&page=&limit=&type=&folder_id=` — filename search
- `GET /media/file?path=` — serve original file (ServeContent for range requests)
- `GET /media/thumb?path=` — serve/generate thumbnail

## Key Architecture Decisions
- No Redux/Zustand/React Query. One context + three hooks.
- No React Router. URL sync via pushState/popstate.
- Server-side deterministic shuffle for pagination stability.
- Path security: Clean -> EvalSymlinks -> filepath.Rel containment check.
- Thumbnail concurrency: singleflight prevents duplicate generation.
- Scan goroutines: context-cancellable, 409 on overlapping rescans.
- No emojis anywhere in the codebase.

## Dev Workflow
```bash
# Terminal 1: Go backend on :8080
go run .
# Terminal 2: Vite dev on :5173 (proxies /api, /media to :8080)
cd frontend && bun run dev
```

## Production Build
```bash
cd frontend && bun run build && cd .. && go build -o local-gallery .
```

## Key Files to Read First
1. `Docs/EndGoal.md` — full specification
2. `Docs/BriefContext.md` — this file
3. `AGENTS.md` — AI agent rules
