package media

import (
	"path/filepath"
	"strings"
	"time"
)

// ImageExtensions defines supported image file extensions.
var ImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
	".avif": true,
	".tiff": true,
}

// VideoExtensions defines supported video file extensions.
var VideoExtensions = map[string]bool{
	".mp4":  true,
	".webm": true,
	".mov":  true,
	".mkv":  true,
	".avi":  true,
	".m4v":  true,
	".ogv":  true,
}

// MimeOverrides provides explicit MIME types for extensions
// where the OS MIME database may be wrong or missing.
var MimeOverrides = map[string]string{
	".avif": "image/avif",
	".webp": "image/webp",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".tiff": "image/tiff",
	".mp4":  "video/mp4",
	".webm": "video/webm",
	".mov":  "video/quicktime",
	".mkv":  "video/x-matroska",
	".avi":  "video/x-msvideo",
	".m4v":  "video/x-m4v",
	".ogv":  "video/ogg",
}

// MediaFile represents a single media file discovered by the scanner.
type MediaFile struct {
	Path      string    `json:"-"`
	Name      string    `json:"-"`
	Extension string    `json:"-"`
	ModTime   time.Time `json:"-"`
	Size      int64     `json:"-"`
	MediaType string    `json:"-"` // "image" or "video"
	FolderID  string    `json:"-"`
}

// IsMedia returns true if the filename has a supported media extension.
func IsMedia(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ImageExtensions[ext] || VideoExtensions[ext]
}

// MediaTypeForExt returns "image", "video", or "" for the given extension.
func MediaTypeForExt(ext string) string {
	ext = strings.ToLower(ext)
	if ImageExtensions[ext] {
		return "image"
	}
	if VideoExtensions[ext] {
		return "video"
	}
	return ""
}

// MimeForExt returns the MIME type for the given extension,
// falling back to application/octet-stream.
func MimeForExt(ext string) string {
	ext = strings.ToLower(ext)
	if mime, ok := MimeOverrides[ext]; ok {
		return mime
	}
	return "application/octet-stream"
}
