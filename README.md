# ZenithLens

ZenithLens is a Linux media gallery for browsing photos and videos stored on
your computer. It runs locally, opens in your browser, and does not upload your
files anywhere.

## What It Does

- Browse images and videos from multiple folders in one gallery
- Pick folders from a native Linux folder dialog
- Search media by filename
- Mark favorites
- Generate and reuse thumbnails for faster browsing
- Keep everything local on your machine

## Requirements

- Linux desktop
- A web browser
- `zenity` recommended for the folder picker
- `kdialog` also supported
- `ffmpeg` optional for better video thumbnails

If neither `zenity` nor `kdialog` is installed, you can still type a folder path
manually.

## Install

Download the latest Linux release package when available, extract it, and run:

```bash
./zenithlens --open
```

This starts ZenithLens and opens it in your browser at:

```text
http://localhost:38471
```

If your system blocks launching the file, make it executable first:

```bash
chmod +x zenithlens
./zenithlens --open
```

## Add Your Media

1. Open ZenithLens.
2. Click `Add Folder`.
3. Click `Choose Folder`.
4. Select a folder from your computer.
5. Click `Add`.

ZenithLens scans that folder and shows supported media files in the gallery.

Folders with spaces in their names are supported.

## Your Data

ZenithLens keeps its own local app data in:

```text
~/.local-gallery/
```

That folder contains:

- `config.json` for registered folders and favorites
- `thumbs/` for cached thumbnails

Your original media files stay where they already are. ZenithLens reads them
from their existing folders.

## Optional System Packages

For the best experience:

- install `zenity` for the folder picker
- install `ffmpeg` for video thumbnails

On KDE systems, `kdialog` can be used instead of `zenity`.

## Troubleshooting

### `Choose Folder` does not open

Install either `zenity` or `kdialog`, then start ZenithLens again.

### Video thumbnails are missing

Install `ffmpeg`. The gallery still works without it, but video thumbnail
generation may be unavailable.

### A folder was added but no media appears

Check that the folder contains supported image or video files and that ZenithLens
has permission to read it.

### I want ZenithLens to open automatically in the browser

Start it with:

```bash
./zenithlens --open
```

## For Developers

Developer setup, architecture, and testing details are documented in:

- `Docs/EndGoal.md`
- `Docs/BriefContext.md`
- `Docs/Testing.md`
