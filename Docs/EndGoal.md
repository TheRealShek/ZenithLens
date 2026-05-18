# ZenithLens: Local Media Gallery

> **Purpose of this file:** Complete implementation spec. An AI reading this file + nothing else should be able to build the entire project from scratch. All architecture decisions are already made — do not re-decide them. Implement exactly what's specified.

---

## Project Overview

Local-only web application that turns arbitrary filesystem folders into a browsable media gallery. Runs as a single Go binary on localhost. No cloud, no auth, no database. Fast, dark, keyboard-friendly.

**Primary use case:** User has photos and videos scattered across many local directories. They register those directories once. The app gives them a unified, visually rich browser interface to explore them all.

---

## Tech Stack — Final, No Debate

| Layer          | Choice                                           | Why                                                             |
| -------------- | ------------------------------------------------ | --------------------------------------------------------------- |
| Backend        | Go (stdlib `net/http`)                           | Filesystem access, range requests for video, no deps to install |
| Frontend       | React + Vite (built to static assets, embedded)  | Component isolation, declarative state, maintainability         |
| Pkg Manager    | Bun                                              | Single binary, fast installs, no nvm/node version management    |
| Thumbnails     | `github.com/disintegration/imaging` (Go)         | Image resize + cache on disk                                    |
| Video thumbs   | `ffmpeg` CLI if available                        | Better UX than browser frame extraction                         |
| Config         | JSON file at `~/.local-gallery/config.json`      | No database. Human-readable.                                    |
| Thumb cache    | `~/.local-gallery/thumbs/`                       | Avoids regenerating. Survives restarts.                         |

---

## Directory Structure

```text
local-gallery/
├── main.go
├── go.mod
├── go.sum
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── scanner/
│   │   └── scanner.go
│   ├── thumb/
│   │   └── thumb.go
│   ├── api/
│   │   ├── handlers.go
│   │   └── dto.go
│   └── media/
│       └── types.go
└── frontend/                     # Vite project root
    ├── index.html                # Vite entry HTML
    ├── package.json
    ├── bun.lock
    ├── vite.config.js
    ├── dist/                     # Build output (Go embeds this)
    └── src/
        ├── main.jsx              # React root mount
        ├── App.jsx               # Top-level layout
        ├── App.css               # Global styles + CSS variables
        ├── api.js                # API fetch wrapper
        ├── hooks/
        │   ├── useRouteState.js   # Route/page/search URL-synced state
        │   ├── useMediaFetch.js   # Data fetching for current view
        │   └── useLightbox.js    # Lightbox open/close/navigate/keyboard
        ├── context/
        │   └── AppContext.jsx    # Shared state (folders, favorites, scanning)
        └── components/
            ├── Sidebar.jsx / .css
            ├── Toolbar.jsx / .css
            ├── Grid.jsx / .css
            ├── MediaItem.jsx / .css
            ├── Pagination.jsx / .css
            ├── Lightbox.jsx / .css
            └── AddFolderModal.jsx / .css
```

Go embeds `frontend/dist/` via `//go:embed frontend/dist`.

---

## Config Schema (`~/.local-gallery/config.json`)

```json
{
  "folders": [
    {
      "id": "uuid-v4",
      "name": "My Photos",
      "path": "/home/user/Pictures",
      "added_at": "2024-01-01T00:00:00Z",
      "last_scanned": "2024-01-01T00:01:00Z",
      "media_count": 1042
    }
  ],
  "favorites": [
    "/home/user/Pictures/vacation/img001.jpg"
  ]
}
```

Config is loaded at startup.

Persistent user state:

* registered folders
* favorites
* metadata

is written back to disk on every mutation.

Runtime-only state:

* scan cache
* pagination state
* in-progress scans
* thumbnail generation locks

exists only in memory.

---

## Media Types

```go
var ImageExtensions = map[string]bool{
    ".jpg": true,
    ".jpeg": true,
    ".png": true,
    ".gif": true,
    ".webp": true,
    ".bmp": true,
    ".avif": true,
    ".tiff": true,
}

var VideoExtensions = map[string]bool{
    ".mp4": true,
    ".webm": true,
    ".mov": true,
    ".mkv": true,
    ".avi": true,
    ".m4v": true,
    ".ogv": true,
}
```

### Important Notes

* HEIC is NOT supported in MVP.
* AVIF support depends on browser support.
* TIFF files may not render inline in all browsers.
* Unsupported formats must still appear in grid with fallback placeholder.

---

## Scanner

```go
func ScanFolder(folderID, folderPath string) ([]MediaFile, error)
func IsMedia(name string) bool
```

### Scanner Rules

* Recursive traversal using `fs.WalkDir`
* Skip hidden files/directories
* Follow normal files only
* Do NOT follow symlinks
* Scan results cached in memory:

```go
map[folderID][]MediaFile
```

### Startup Behavior

Startup scans happen asynchronously.

During initial scan:

```json
{
  "scanning": true
}
```

must be included in paginated API responses.

Frontend must treat totals/pages as unstable while scanning.

---

## API Specification

Base:

```text
http://localhost:8080
```

---

## Folder Endpoints

| Method | Path                      |
| ------ | ------------------------- |
| GET    | `/api/folders`            |
| POST   | `/api/folders`            |
| DELETE | `/api/folders/:id`        |
| POST   | `/api/folders/:id/rescan` |

---

## Media Endpoints

| Method | Path                                     |
| ------ | ---------------------------------------- |
| GET    | `/api/home`                              |
| GET    | `/api/folder/:id`                        |
| GET    | `/api/favorites`                         |
| POST   | `/api/favorites`                         |
| DELETE | `/api/favorites?path=<encoded_abs_path>` |
| GET    | `/api/search`                            |

---

## Search Endpoint

```http
GET /api/search?q=cat&page=1&limit=50&type=image&folder_id=abc
```

### Search Params

| Param       | Default  |
| ----------- | -------- |
| `q`         | required |
| `page`      | 1        |
| `limit`     | 50       |
| `type`      | all      |
| `folder_id` | optional |

### Search Rules

* filename-only search
* case-insensitive
* substring match
* paginated
* stable ordering within same seed

---

## File Serving

| Method | Path           |
| ------ | -------------- |
| GET    | `/media/file`  |
| GET    | `/media/thumb` |

### File Validation

All requested paths MUST:

* resolve symlinks using `filepath.EvalSymlinks`
* normalize with `filepath.Clean`
* remain inside one registered folder

Preferred validation:

```go
filepath.Rel(...)
```

Return:

* `403` on traversal attempt
* `404` if file missing

---

## MIME Types

Do NOT rely entirely on OS MIME database.

Explicit MIME overrides required:

```go
.avif -> image/avif
.webp -> image/webp
```

Fallback:

```text
application/octet-stream
```

---

## Thumbnail Generation

```go
const ThumbWidth = 400
```

### Cache Key

Do NOT hash path only.

Use:

```go
sha256(path + modtime + filesize)
```

to invalidate stale thumbnails automatically.

---

## Thumbnail Concurrency

Thumbnail generation MUST use request coalescing.

Example acceptable approaches:

* `singleflight.Group`
* keyed mutex map

Goal:

* prevent duplicate concurrent thumbnail generation for same file

---

## Video Thumbnail Generation

```bash
ffmpeg -ss 1 -i <input> -vframes 1 -q:v 2 -vf scale=400:-1 <output.jpg>
```

### ffmpeg Rules

* MUST execute with timeout using `exec.CommandContext`
* corrupt files must fail gracefully
* timeout failure returns empty thumb
* app continues functioning without ffmpeg

If ffmpeg unavailable:

* frontend falls back to `<video>` preview

---

## Frontend Specification

React + Vite frontend built to static assets and embedded into the Go binary.

Development uses Vite dev server on `:5173` with proxy to Go backend on `:8080`.
Production build outputs to `frontend/dist/`, which Go embeds and serves via `http.FileServer`.

No SSR. No React Router (URL state managed via `pushState`/`popstate`). No Redux/Zustand/React Query.

State architecture:

* `AppContext` — shared state: folders, favorites, scanning status, polling
* `useRouteState` hook — URL-synced route, page, seed, folderId, searchQuery
* `useMediaFetch` hook — fetches items/total for current route+page
* `useLightbox` hook — lightbox open/close, index, keyboard shortcuts

CSS: co-located `.css` files per component. Global variables in `App.css`.

---

## Unsupported Media UX

If browser cannot render media:

* show fallback placeholder
* show file extension badge
* allow download/open externally
* never render broken-image icon silently

---

## Pagination

Server-side deterministic shuffle:

```go
func shuffleWithSeed(files []MediaFile, seed int64) []MediaFile
```

### Important Behavior

Seed guarantees stable ordering ONLY for same dataset snapshot.

After rescans/additions/removals:

* pagination order may change
* totals may change
* page contents may shift

This is expected behavior.

---

## Lightbox

Keyboard shortcuts:

| Key | Action          |
| --- | --------------- |
| ←   | Previous        |
| →   | Next            |
| Esc | Close           |
| F   | Toggle favorite |
| D   | Download        |

---

## Security Constraints

### MUST Defend Against

* path traversal
* symlink escape
* malformed query paths

### MUST NOT

* expose arbitrary filesystem paths
* trust raw query params directly

---

## Startup Behavior

```text
$ ./local-gallery
→ Load config
→ Create thumb cache dir
→ Start async folder scans
→ Listen on :8080
→ Print localhost URL
```

Optional:

```bash
--open
```

opens browser automatically.

---

## Performance Notes

Accepted trade-offs:

* startup scan cost acceptable
* in-memory cache acceptable
* no SQLite/indexer intentionally

Known scaling cost:

* deterministic shuffle copies slice O(n)
* large libraries consume RAM proportional to media count

This is acceptable for MVP scope.

---

## Go Module

```go
module github.com/therealshek/local-gallery

go 1.22

require (
    github.com/disintegration/imaging v1.6.2
)
```

---

## Implementation Order

1. media types
2. config
3. scanner
4. thumbnails
5. API handlers
6. main wiring
7. frontend scaffold (Vite + React + proxy)
8. frontend components (sidebar, grid, pagination, lightbox)
9. production build + Go embed integration

Build/test after every step.

---

## Non-Goals

* no auth
* no cloud sync
* no mobile responsiveness
* no transcoding
* no metadata editing
* no database/indexer

---

## Final Design Principles

* local-first
* single binary deployment (frontend embedded)
* desktop-focused
* visually dense
* deterministic pagination
* graceful degradation over hard failure
* simple deployment over maximum scalability
* frontend: minimal abstractions, no heavy state libraries
