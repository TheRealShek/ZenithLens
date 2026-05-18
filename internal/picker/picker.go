package picker

import (
	"errors"
	"os/exec"
	"strings"
)

var (
	ErrUnavailable = errors.New("no supported folder picker available")
	ErrCancelled   = errors.New("folder selection cancelled")
)

// PickFolder opens the first supported Linux folder picker and returns the selected path.
func PickFolder() (string, error) {
	if path, err := exec.LookPath("zenity"); err == nil {
		return run(path, "--file-selection", "--directory")
	}
	if path, err := exec.LookPath("kdialog"); err == nil {
		return run(path, "--getexistingdirectory")
	}
	return "", ErrUnavailable
}

func run(command string, args ...string) (string, error) {
	out, err := exec.Command(command, args...).Output()
	if err != nil {
		return "", ErrCancelled
	}
	selected := strings.TrimRight(string(out), "\r\n")
	if selected == "" {
		return "", ErrCancelled
	}
	return selected, nil
}
