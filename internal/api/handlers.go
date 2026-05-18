package api

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/therealshek/local-gallery/internal/config"
	"github.com/therealshek/local-gallery/internal/media"
	"github.com/therealshek/local-gallery/internal/picker"
	"github.com/therealshek/local-gallery/internal/scanner"
	"github.com/therealshek/local-gallery/internal/thumb"
)

// Handler holds all runtime state and serves API requests.
type Handler struct {
	Config      *config.Config
	ConfigMutex sync.RWMutex

	MediaCache      map[string][]media.MediaFile
	ScanningFolders map[string]bool
	ScanCancels     map[string]context.CancelFunc
	CacheMutex      sync.RWMutex

	ThumbDir  string
	HasFFmpeg bool
}

// NewHandler creates a Handler with initialized maps.
func NewHandler(cfg *config.Config, thumbDir string, hasFFmpeg bool) *Handler {
	return &Handler{
		Config:          cfg,
		MediaCache:      make(map[string][]media.MediaFile),
		ScanningFolders: make(map[string]bool),
		ScanCancels:     make(map[string]context.CancelFunc),
		ThumbDir:        thumbDir,
		HasFFmpeg:       hasFFmpeg,
	}
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a standard error JSON response.
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// --- Scan helpers ---

// StartScan launches an async scan for a folder.
func (h *Handler) StartScan(folderID, folderPath string) {
	ctx, cancel := context.WithCancel(context.Background())

	h.CacheMutex.Lock()
	h.ScanningFolders[folderID] = true
	h.ScanCancels[folderID] = cancel
	h.CacheMutex.Unlock()

	go func() {
		files, err := scanner.ScanFolder(ctx, folderID, folderPath)

		h.CacheMutex.Lock()
		if ctx.Err() == nil && err == nil {
			h.MediaCache[folderID] = files
		}
		delete(h.ScanningFolders, folderID)
		delete(h.ScanCancels, folderID)
		h.CacheMutex.Unlock()

		if ctx.Err() == nil && err == nil {
			h.ConfigMutex.Lock()
			if f := h.Config.FindFolder(folderID); f != nil {
				f.MediaCount = len(files)
				f.LastScanned = time.Now()
				config.Save(h.Config)
			}
			h.ConfigMutex.Unlock()
		}
	}()
}

// CancelAllScans cancels all active scans for shutdown.
func (h *Handler) CancelAllScans() {
	h.CacheMutex.Lock()
	defer h.CacheMutex.Unlock()
	for _, cancel := range h.ScanCancels {
		cancel()
	}
}

// isScanning returns true if any folder is currently scanning.
func (h *Handler) isScanning() bool {
	for _, v := range h.ScanningFolders {
		if v {
			return true
		}
	}
	return false
}

// --- Pagination helpers ---

func shuffleWithSeed(files []media.MediaFile, seed int64) []media.MediaFile {
	out := make([]media.MediaFile, len(files))
	copy(out, files)
	r := rand.New(rand.NewSource(seed))
	r.Shuffle(len(out), func(i, j int) {
		out[i], out[j] = out[j], out[i]
	})
	return out
}

func paginateFiles(files []media.MediaFile, page, limit int) []media.MediaFile {
	start := (page - 1) * limit
	if start >= len(files) {
		return nil
	}
	end := start + limit
	if end > len(files) {
		end = len(files)
	}
	return files[start:end]
}

func toDTO(files []media.MediaFile) []MediaFileDTO {
	dtos := make([]MediaFileDTO, len(files))
	for i, f := range files {
		dtos[i] = MediaFileDTO{
			Path:     f.Path,
			Name:     f.Name,
			Type:     f.MediaType,
			MimeType: media.MimeForExt(f.Extension),
		}
	}
	return dtos
}

func parsePageLimit(r *http.Request) (int, int) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return page, limit
}

func parseSeed(r *http.Request) int64 {
	s, err := strconv.ParseInt(r.URL.Query().Get("seed"), 10, 64)
	if err != nil || s == 0 {
		return time.Now().UnixNano()
	}
	return s
}

// --- Folder Endpoints ---

// GetFolders handles GET /api/folders
func (h *Handler) GetFolders(w http.ResponseWriter, r *http.Request) {
	h.ConfigMutex.RLock()
	folders := h.Config.Folders
	h.ConfigMutex.RUnlock()

	h.CacheMutex.RLock()
	defer h.CacheMutex.RUnlock()

	dtos := make([]FolderDTO, len(folders))
	for i, f := range folders {
		dtos[i] = FolderDTO{
			ID:          f.ID,
			Name:        f.Name,
			Path:        f.Path,
			AddedAt:     f.AddedAt.Format(time.RFC3339),
			LastScanned: f.LastScanned.Format(time.RFC3339),
			MediaCount:  f.MediaCount,
			Scanning:    h.ScanningFolders[f.ID],
		}
	}
	writeJSON(w, http.StatusOK, dtos)
}

// AddFolder handles POST /api/folders
func (h *Handler) AddFolder(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if body.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	// Validate path exists and is a directory.
	info, err := os.Stat(body.Path)
	if err != nil || !info.IsDir() {
		writeError(w, http.StatusBadRequest, "path does not exist or is not a directory")
		return
	}

	absPath, err := filepath.Abs(body.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid path")
		return
	}

	name := body.Name
	if name == "" {
		name = filepath.Base(absPath)
	}

	folder := config.Folder{
		ID:      uuid.New().String(),
		Name:    name,
		Path:    absPath,
		AddedAt: time.Now(),
	}

	h.ConfigMutex.Lock()
	h.Config.Folders = append(h.Config.Folders, folder)
	config.Save(h.Config)
	h.ConfigMutex.Unlock()

	h.StartScan(folder.ID, folder.Path)

	writeJSON(w, http.StatusCreated, FolderDTO{
		ID:      folder.ID,
		Name:    folder.Name,
		Path:    folder.Path,
		AddedAt: folder.AddedAt.Format(time.RFC3339),
	})
}

// PickFolder handles POST /api/folders/pick.
func (h *Handler) PickFolder(w http.ResponseWriter, r *http.Request) {
	path, err := picker.PickFolder()
	if err != nil {
		switch err {
		case picker.ErrUnavailable:
			writeError(w, http.StatusServiceUnavailable, "no supported folder picker available")
		case picker.ErrCancelled:
			writeError(w, http.StatusConflict, "folder selection cancelled")
		default:
			writeError(w, http.StatusInternalServerError, "folder selection failed")
		}
		return
	}

	writeJSON(w, http.StatusOK, PickFolderDTO{Path: path})
}

// DeleteFolder handles DELETE /api/folders/:id
func (h *Handler) DeleteFolder(w http.ResponseWriter, r *http.Request, id string) {
	// Cancel active scan if any.
	h.CacheMutex.Lock()
	if cancel, ok := h.ScanCancels[id]; ok {
		cancel()
	}
	delete(h.MediaCache, id)
	delete(h.ScanningFolders, id)
	delete(h.ScanCancels, id)
	h.CacheMutex.Unlock()

	h.ConfigMutex.Lock()
	if !h.Config.RemoveFolder(id) {
		h.ConfigMutex.Unlock()
		writeError(w, http.StatusNotFound, "folder not found")
		return
	}
	config.Save(h.Config)
	h.ConfigMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// RescanFolder handles POST /api/folders/:id/rescan
func (h *Handler) RescanFolder(w http.ResponseWriter, r *http.Request, id string) {
	h.CacheMutex.RLock()
	if h.ScanningFolders[id] {
		h.CacheMutex.RUnlock()
		writeError(w, http.StatusConflict, "scan already in progress")
		return
	}
	h.CacheMutex.RUnlock()

	h.ConfigMutex.RLock()
	folder := h.Config.FindFolder(id)
	h.ConfigMutex.RUnlock()

	if folder == nil {
		writeError(w, http.StatusNotFound, "folder not found")
		return
	}

	h.StartScan(folder.ID, folder.Path)
	writeJSON(w, http.StatusOK, map[string]string{"status": "scanning"})
}

// --- Media Endpoints ---

// GetHome handles GET /api/home
func (h *Handler) GetHome(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePageLimit(r)
	seed := parseSeed(r)

	h.CacheMutex.RLock()
	scanning := h.isScanning()
	var all []media.MediaFile
	for _, files := range h.MediaCache {
		all = append(all, files...)
	}
	h.CacheMutex.RUnlock()

	shuffled := shuffleWithSeed(all, seed)
	paged := paginateFiles(shuffled, page, limit)

	writeJSON(w, http.StatusOK, PageResponseDTO{
		Items:    toDTO(paged),
		Total:    len(all),
		Page:     page,
		Scanning: scanning,
	})
}

// GetFolderMedia handles GET /api/folder/:id
func (h *Handler) GetFolderMedia(w http.ResponseWriter, r *http.Request, id string) {
	page, limit := parsePageLimit(r)
	seed := parseSeed(r)

	h.CacheMutex.RLock()
	files, ok := h.MediaCache[id]
	scanning := h.ScanningFolders[id]
	h.CacheMutex.RUnlock()

	if !ok {
		h.ConfigMutex.RLock()
		folder := h.Config.FindFolder(id)
		h.ConfigMutex.RUnlock()
		if folder == nil {
			writeError(w, http.StatusNotFound, "folder not found")
			return
		}
		// Folder exists but hasn't been scanned yet.
		files = nil
	}

	shuffled := shuffleWithSeed(files, seed)
	paged := paginateFiles(shuffled, page, limit)

	writeJSON(w, http.StatusOK, PageResponseDTO{
		Items:    toDTO(paged),
		Total:    len(files),
		Page:     page,
		Scanning: scanning,
	})
}

// GetFavorites handles GET /api/favorites
func (h *Handler) GetFavorites(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePageLimit(r)

	h.ConfigMutex.RLock()
	favPaths := make(map[string]bool, len(h.Config.Favorites))
	for _, p := range h.Config.Favorites {
		favPaths[p] = true
	}
	h.ConfigMutex.RUnlock()

	h.CacheMutex.RLock()
	var favFiles []media.MediaFile
	for _, files := range h.MediaCache {
		for _, f := range files {
			if favPaths[f.Path] {
				favFiles = append(favFiles, f)
			}
		}
	}
	h.CacheMutex.RUnlock()

	paged := paginateFiles(favFiles, page, limit)

	writeJSON(w, http.StatusOK, PageResponseDTO{
		Items:    toDTO(paged),
		Total:    len(favFiles),
		Page:     page,
		Scanning: false,
	})
}

// AddFavorite handles POST /api/favorites
func (h *Handler) AddFavorite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Path string `json:"path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	h.ConfigMutex.Lock()
	h.Config.AddFavorite(body.Path)
	config.Save(h.Config)
	h.ConfigMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// RemoveFavorite handles DELETE /api/favorites?path=...
func (h *Handler) RemoveFavorite(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		writeError(w, http.StatusBadRequest, "path query param is required")
		return
	}

	h.ConfigMutex.Lock()
	h.Config.RemoveFavorite(path)
	config.Save(h.Config)
	h.ConfigMutex.Unlock()

	w.WriteHeader(http.StatusNoContent)
}

// Search handles GET /api/search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q is required")
		return
	}
	page, limit := parsePageLimit(r)
	seed := parseSeed(r)
	typeFilter := r.URL.Query().Get("type")
	folderID := r.URL.Query().Get("folder_id")

	qLower := strings.ToLower(q)

	h.CacheMutex.RLock()
	scanning := h.isScanning()
	var matches []media.MediaFile
	for fid, files := range h.MediaCache {
		if folderID != "" && fid != folderID {
			continue
		}
		for _, f := range files {
			if typeFilter != "" && typeFilter != "all" && f.MediaType != typeFilter {
				continue
			}
			if strings.Contains(strings.ToLower(f.Name), qLower) {
				matches = append(matches, f)
			}
		}
	}
	h.CacheMutex.RUnlock()

	shuffled := shuffleWithSeed(matches, seed)
	paged := paginateFiles(shuffled, page, limit)

	writeJSON(w, http.StatusOK, PageResponseDTO{
		Items:    toDTO(paged),
		Total:    len(matches),
		Page:     page,
		Scanning: scanning,
	})
}

// --- File Serving ---

// validatePath validates and resolves a media path, returning the resolved path
// or an error code.
func (h *Handler) validatePath(rawPath string) (string, int) {
	if rawPath == "" {
		return "", http.StatusBadRequest
	}

	cleaned := filepath.Clean(rawPath)
	resolved, err := filepath.EvalSymlinks(cleaned)
	if err != nil {
		return "", http.StatusNotFound
	}

	h.ConfigMutex.RLock()
	folders := h.Config.Folders
	h.ConfigMutex.RUnlock()

	contained := false
	for _, folder := range folders {
		folderResolved, err := filepath.EvalSymlinks(folder.Path)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(folderResolved, resolved)
		if err != nil {
			continue
		}
		if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		contained = true
		break
	}

	if !contained {
		return "", http.StatusForbidden
	}

	return resolved, 0
}

// ServeMediaFile handles GET /media/file
func (h *Handler) ServeMediaFile(w http.ResponseWriter, r *http.Request) {
	rawPath := r.URL.Query().Get("path")
	resolved, errCode := h.validatePath(rawPath)
	if errCode != 0 {
		if errCode == http.StatusForbidden {
			writeError(w, errCode, "access denied")
		} else {
			writeError(w, errCode, "not found")
		}
		return
	}

	stat, err := os.Stat(resolved)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	f, err := os.Open(resolved)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}
	defer f.Close()

	ext := strings.ToLower(filepath.Ext(resolved))
	w.Header().Set("Content-Type", media.MimeForExt(ext))
	http.ServeContent(w, r, f.Name(), stat.ModTime(), f)
}

// ServeThumb handles GET /media/thumb
func (h *Handler) ServeThumb(w http.ResponseWriter, r *http.Request) {
	rawPath := r.URL.Query().Get("path")
	resolved, errCode := h.validatePath(rawPath)
	if errCode != 0 {
		if errCode == http.StatusForbidden {
			writeError(w, errCode, "access denied")
		} else {
			writeError(w, errCode, "not found")
		}
		return
	}

	stat, err := os.Stat(resolved)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	thumb.Serve(w, r, resolved, stat.ModTime(), stat.Size(), h.ThumbDir, h.HasFFmpeg)
}

// --- Router ---

// RegisterRoutes sets up all routes on the given mux.
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/folders/pick", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.PickFolder(w, r)
	})

	mux.HandleFunc("/api/folders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetFolders(w, r)
		case http.MethodPost:
			h.AddFolder(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/api/folders/", func(w http.ResponseWriter, r *http.Request) {
		// Parse: /api/folders/:id or /api/folders/:id/rescan
		path := strings.TrimPrefix(r.URL.Path, "/api/folders/")
		parts := strings.SplitN(path, "/", 2)
		id := parts[0]
		if id == "" {
			writeError(w, http.StatusBadRequest, "folder id required")
			return
		}

		if len(parts) == 2 && parts[1] == "rescan" {
			if r.Method != http.MethodPost {
				writeError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.RescanFolder(w, r, id)
			return
		}

		if r.Method == http.MethodDelete {
			h.DeleteFolder(w, r, id)
			return
		}

		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	})

	mux.HandleFunc("/api/home", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetHome(w, r)
	})

	mux.HandleFunc("/api/folder/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/folder/")
		if id == "" {
			writeError(w, http.StatusBadRequest, "folder id required")
			return
		}
		h.GetFolderMedia(w, r, id)
	})

	mux.HandleFunc("/api/favorites", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.GetFavorites(w, r)
		case http.MethodPost:
			h.AddFavorite(w, r)
		case http.MethodDelete:
			h.RemoveFavorite(w, r)
		default:
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.Search(w, r)
	})

	mux.HandleFunc("/media/file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.ServeMediaFile(w, r)
	})

	mux.HandleFunc("/media/thumb", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.ServeThumb(w, r)
	})
}
