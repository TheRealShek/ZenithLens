package scanner

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/therealshek/local-gallery/internal/media"
)

// ScanFolder walks folderPath recursively and returns all media files found.
// It respects context cancellation for clean shutdown.
func ScanFolder(ctx context.Context, folderID, folderPath string) ([]media.MediaFile, error) {
	var files []media.MediaFile

	err := filepath.WalkDir(folderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}

		// Check for cancellation.
		if ctx.Err() != nil {
			return filepath.SkipAll
		}

		name := d.Name()

		// Skip hidden files and directories.
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories and non-regular files (symlinks).
		if d.IsDir() {
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}

		if !media.IsMedia(name) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil // skip files we can't stat
		}

		ext := strings.ToLower(filepath.Ext(name))
		files = append(files, media.MediaFile{
			Path:      path,
			Name:      name,
			Extension: ext,
			ModTime:   info.ModTime(),
			Size:      info.Size(),
			MediaType: media.MediaTypeForExt(ext),
			FolderID:  folderID,
		})

		return nil
	})

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return files, err
}
