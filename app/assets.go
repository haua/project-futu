package app

import (
	"embed"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
)

//go:embed embedded_assets/icon.png embedded_assets/icon-edit.png
var embeddedAssets embed.FS

var embeddedAssetPaths = map[string]string{
	"icon.png":      "embedded_assets/icon.png",
	"icon-edit.png": "embedded_assets/icon-edit.png",
}

func loadEmbeddedAssetResource(fileName string) fyne.Resource {
	path, ok := embeddedAssetPaths[fileName]
	if !ok {
		return nil
	}
	content, err := embeddedAssets.ReadFile(path)
	if err != nil {
		return nil
	}
	return fyne.NewStaticResource(fileName, content)
}

func loadAssetResourceByName(fileName string) fyne.Resource {
	if res := loadEmbeddedAssetResource(fileName); res != nil {
		return res
	}

	candidates := []string{
		filepath.Join("assets", fileName),
	}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append([]string{filepath.Join(exeDir, "assets", fileName)}, candidates...)
	}

	for _, path := range candidates {
		content, readErr := os.ReadFile(path)
		if readErr == nil {
			return fyne.NewStaticResource(fileName, content)
		}
	}

	return nil
}
