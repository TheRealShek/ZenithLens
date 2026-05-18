# ZenithLens Testing Guide

This guide covers the current Linux-native workflow: backend checks, frontend
build checks, manual browser testing, folder picker behavior, media serving, and
persistence across restarts.

## Prerequisites

Install or confirm:

- Go 1.22+
- Bun
- a Linux desktop session
- `zenity` recommended, or `kdialog` as fallback
- `ffmpeg` optional if you want to verify video thumbnails

Useful checks:

```bash
go version
bun --version
command -v zenity || command -v kdialog
command -v ffmpeg
```

Create or choose one test media folder on the host with:

- at least one image
- optionally one video
- optionally a folder name containing spaces, such as:

```text
/home/user/Test Media
```

## 1. Run Automated Backend Checks

From the repo root:

```bash
go test ./...
go vet ./...
```

Expected result:

- all Go packages pass
- no vet findings

If your environment has a read-only default Go cache, use:

```bash
GOCACHE=/tmp/zenithlens-go-cache go test ./...
```

## 2. Build the Frontend

From the repo root:

```bash
cd frontend
bun install
bun run build
cd ..
```

Expected result:

- Vite completes successfully
- `frontend/dist/` is created or refreshed

## 3. Run the App in Development

Terminal 1 from the repo root:

```bash
go run .
```

Expected backend output:

```text
ZenithLens running at http://localhost:38471
```

Terminal 2:

```bash
cd frontend
bun run dev
```

Open the frontend dev URL printed by Vite, usually:

```text
http://localhost:5173
```

The Vite dev server proxies `/api` and `/media` to the Go backend.

## 4. Verify Folder Picker Flow

1. Open `Add Folder`.
2. Click `Choose Folder`.
3. Confirm that a native folder dialog opens.
4. Select your test media folder.
5. Confirm that the absolute Linux path appears in the input field.
6. Click `Add`.

Expected result:

- the folder appears in the sidebar
- scanning begins and completes
- media from the selected host folder appears in the gallery

Folder names with spaces should work when selected through the picker because the
backend receives one full path string, not a shell-split command.

## 5. Verify Manual Path Fallback

To test the fallback manually, use a session where neither picker is available,
or temporarily run the backend with a restricted `PATH` that does not include
`zenity` or `kdialog`.

Then:

1. Open `Add Folder`.
2. Click `Choose Folder`.
3. Confirm that the UI shows a picker-unavailable error.
4. Enter the absolute host path manually.
5. Click `Add`.

Expected result:

- the folder still registers and scans successfully

## 6. Verify Gallery Behavior

After a folder is registered:

1. Confirm thumbnails load in the grid.
2. Open an image or video in the lightbox.
3. Test:
   - previous / next navigation
   - favorite toggle
   - search by filename
   - folder-specific view
   - pagination if enough items exist
4. If video files are present:
   - with `ffmpeg` installed, confirm video thumbnails generate
   - without `ffmpeg`, confirm the app still functions

## 7. Verify Persistence Across Restart

1. Add at least one folder.
2. Favorite at least one item.
3. Stop the backend with `Ctrl+C`.
4. Start it again:

```bash
go run .
```

5. Refresh the UI.

Expected result:

- previously registered folders are still present
- favorites are still present
- thumbnail cache remains under `~/.local-gallery/thumbs/`
- startup rescans begin automatically

You can inspect durable state directly:

```bash
ls ~/.local-gallery
cat ~/.local-gallery/config.json
```

## 8. Verify Production Binary

From the repo root:

```bash
cd frontend
bun run build
cd ..
go build -o zenithlens .
./zenithlens --open
```

Expected result:

- the browser opens to `http://localhost:38471`
- the embedded frontend loads without the Vite dev server
- folder picker and gallery behavior still work

## Suggested Test Matrix

Run at least these cases before release:

| Area | Cases |
| --- | --- |
| Picker | `zenity`, `kdialog`, neither available, user cancel |
| Paths | normal path, path containing spaces |
| Media | image-only folder, mixed image/video folder |
| Persistence | restart after folders + favorites exist |
| Thumbnails | with `ffmpeg`, without `ffmpeg` |
| Build | dev mode, production binary |

## Common Failures

| Symptom | Likely Cause |
| --- | --- |
| `Choose Folder` returns an error | no `zenity` or `kdialog`, or no GUI session |
| folder add returns `path does not exist` | typo or inaccessible host path |
| video thumbnails missing | `ffmpeg` not installed or source video unsupported |
| frontend changes not visible in binary | frontend was not rebuilt before `go build` |
