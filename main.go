package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/therealshek/local-gallery/internal/api"
	"github.com/therealshek/local-gallery/internal/config"
)

//go:embed frontend/dist
var frontendFS embed.FS

func main() {
	openBrowser := flag.Bool("open", false, "open browser automatically")
	flag.Parse()

	// Load config.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create thumb cache directory.
	thumbDir := config.ThumbDir()
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		log.Fatalf("failed to create thumb dir: %v", err)
	}

	// Detect ffmpeg.
	_, ffmpegErr := exec.LookPath("ffmpeg")
	hasFFmpeg := ffmpegErr == nil

	// Create handler.
	handler := api.NewHandler(cfg, thumbDir, hasFFmpeg)

	// Register routes.
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Serve frontend (Vite build output).
	sub, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		log.Fatalf("failed to create sub filesystem: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(sub)))

	// Start async scans for all registered folders.
	for _, folder := range cfg.Folders {
		handler.StartScan(folder.ID, folder.Path)
	}

	// Handle shutdown.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nShutting down...")
		handler.CancelAllScans()
		os.Exit(0)
	}()

	addr := ":38471"
	url := "http://localhost" + addr
	fmt.Printf("ZenithLens running at %s\n", url)

	if *openBrowser {
		exec.Command("xdg-open", url).Start()
	}

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
