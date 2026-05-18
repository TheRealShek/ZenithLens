package api

// FolderDTO is the JSON representation of a registered folder.
type FolderDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	AddedAt     string `json:"added_at"`
	LastScanned string `json:"last_scanned"`
	MediaCount  int    `json:"media_count"`
	Scanning    bool   `json:"scanning"`
}

// PickFolderDTO is returned after the user selects a folder in the native picker.
type PickFolderDTO struct {
	Path string `json:"path"`
}

// MediaFileDTO is the JSON representation of a single media file.
type MediaFileDTO struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Type     string `json:"type"` // "image" or "video"
	MimeType string `json:"mimeType"`
}

// PageResponseDTO is the paginated response for media listing endpoints.
type PageResponseDTO struct {
	Items    []MediaFileDTO `json:"items"`
	Total    int            `json:"total"`
	Page     int            `json:"page"`
	Scanning bool           `json:"scanning"`
}
