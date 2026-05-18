package thumb

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"github.com/therealshek/local-gallery/internal/media"
	"golang.org/x/sync/singleflight"
)

const ThumbWidth = 400

var group singleflight.Group

// CacheKey computes a deterministic cache key from path, modtime, and size.
func CacheKey(path string, modTime time.Time, size int64) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s|%d|%d", path, modTime.UnixNano(), size)
	return fmt.Sprintf("%x", h.Sum(nil))
}

// CachePath returns the full path to the cached thumbnail.
func CachePath(thumbDir, key string) string {
	return filepath.Join(thumbDir, key+".jpg")
}

// Serve generates (if needed) and serves a thumbnail via http.ServeContent.
// hasFFmpeg indicates whether ffmpeg is available on the system.
func Serve(w http.ResponseWriter, r *http.Request, filePath string, modTime time.Time, size int64, thumbDir string, hasFFmpeg bool) {
	ext := strings.ToLower(filepath.Ext(filePath))
	mediaType := media.MediaTypeForExt(ext)

	key := CacheKey(filePath, modTime, size)
	cached := CachePath(thumbDir, key)

	// Check cache hit.
	if st, err := os.Stat(cached); err == nil {
		f, err := os.Open(cached)
		if err == nil {
			defer f.Close()
			w.Header().Set("Content-Type", "image/jpeg")
			http.ServeContent(w, r, st.Name(), st.ModTime(), f)
			return
		}
	}

	// Generate via singleflight.
	_, err, _ := group.Do(key, func() (interface{}, error) {
		return nil, generate(filePath, cached, mediaType, thumbDir, hasFFmpeg)
	})
	if err != nil {
		http.Error(w, "thumbnail generation failed", http.StatusNotFound)
		return
	}

	f, err := os.Open(cached)
	if err != nil {
		http.Error(w, "thumbnail not found", http.StatusNotFound)
		return
	}
	defer f.Close()
	st, _ := f.Stat()
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, st.Name(), st.ModTime(), f)
}

func generate(filePath, cachePath, mediaType, thumbDir string, hasFFmpeg bool) error {
	// Create temp file in the thumbs directory.
	tmp, err := os.CreateTemp(thumbDir, "thumb-*.jpg")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	tmp.Close()

	success := false
	defer func() {
		if !success {
			os.Remove(tmpName)
		}
	}()

	if mediaType == "image" {
		return generateImage(filePath, tmpName, cachePath, &success)
	}
	if mediaType == "video" {
		if !hasFFmpeg {
			return fmt.Errorf("ffmpeg not available")
		}
		return generateVideo(filePath, tmpName, cachePath, &success)
	}
	return fmt.Errorf("unsupported media type")
}

func generateImage(filePath, tmpPath, cachePath string, success *bool) error {
	img, err := imaging.Open(filePath)
	if err != nil {
		return err
	}
	resized := imaging.Resize(img, ThumbWidth, 0, imaging.Lanczos)
	if err := imaging.Save(resized, tmpPath); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, cachePath); err != nil {
		return err
	}
	*success = true
	return nil
}

func generateVideo(filePath, tmpPath, cachePath string, success *bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", "1",
		"-i", filePath,
		"-vframes", "1",
		"-q:v", "2",
		"-vf", fmt.Sprintf("scale=%d:-1", ThumbWidth),
		"-y",
		tmpPath,
	)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, cachePath); err != nil {
		return err
	}
	*success = true
	return nil
}
