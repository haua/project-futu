package utils

import (
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

// LoadAssetResource loads an icon resource from assets under exe dir or cwd.
func LoadAssetResource(fileName string) fyne.Resource {
	exePath, err := os.Executable()
	if err != nil {
		return nil
	}

	exeDir := filepath.Dir(exePath)
	candidates := []string{
		filepath.Join(exeDir, "assets", fileName),
		filepath.Join("assets", fileName),
	}

	for _, path := range candidates {
		content, readErr := os.ReadFile(path)
		if readErr == nil {
			return fyne.NewStaticResource(fileName, content)
		}
	}

	return nil
}
