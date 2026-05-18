package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Folder represents a registered media folder.
type Folder struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	AddedAt     time.Time `json:"added_at"`
	LastScanned time.Time `json:"last_scanned"`
	MediaCount  int       `json:"media_count"`
}

// Config holds all persistent user state.
type Config struct {
	Folders   []Folder `json:"folders"`
	Favorites []string `json:"favorites"`
}

// DefaultDir returns the path to ~/.local-gallery/.
func DefaultDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local-gallery")
}

// DefaultPath returns the path to ~/.local-gallery/config.json.
func DefaultPath() string {
	return filepath.Join(DefaultDir(), "config.json")
}

// ThumbDir returns the path to ~/.local-gallery/thumbs/.
func ThumbDir() string {
	return filepath.Join(DefaultDir(), "thumbs")
}

// Load reads the config from disk. If the file does not exist,
// returns an empty config.
func Load() (*Config, error) {
	path := DefaultPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Folders:   []Folder{},
				Favorites: []string{},
			}, nil
		}
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Folders == nil {
		cfg.Folders = []Folder{}
	}
	if cfg.Favorites == nil {
		cfg.Favorites = []string{}
	}
	return &cfg, nil
}

// Save writes the config to disk atomically.
func Save(cfg *Config) error {
	dir := DefaultDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := DefaultPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, DefaultPath())
}

// FindFolder returns the folder with the given ID, or nil.
func (c *Config) FindFolder(id string) *Folder {
	for i := range c.Folders {
		if c.Folders[i].ID == id {
			return &c.Folders[i]
		}
	}
	return nil
}

// RemoveFolder removes the folder with the given ID. Returns true if found.
func (c *Config) RemoveFolder(id string) bool {
	for i := range c.Folders {
		if c.Folders[i].ID == id {
			c.Folders = append(c.Folders[:i], c.Folders[i+1:]...)
			return true
		}
	}
	return false
}

// IsFavorite returns true if the path is in the favorites list.
func (c *Config) IsFavorite(path string) bool {
	for _, f := range c.Favorites {
		if f == path {
			return true
		}
	}
	return false
}

// AddFavorite adds the path to favorites if not already present.
func (c *Config) AddFavorite(path string) {
	if !c.IsFavorite(path) {
		c.Favorites = append(c.Favorites, path)
	}
}

// RemoveFavorite removes the path from favorites.
func (c *Config) RemoveFavorite(path string) {
	for i, f := range c.Favorites {
		if f == path {
			c.Favorites = append(c.Favorites[:i], c.Favorites[i+1:]...)
			return
		}
	}
}
