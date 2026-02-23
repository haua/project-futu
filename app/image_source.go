package app

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	imageSourceModeKey      = "image.source_mode"
	fixedImagePathPrefKey   = "image.fixed_path"
	randomFolderPathPrefKey = "image.random_folder"
	imageSourceModeSingle   = "single"
	imageSourceModeFolder   = "folder"
)

func normalizeImageSourceMode(mode string) string {
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case imageSourceModeFolder:
		return imageSourceModeFolder
	default:
		return imageSourceModeSingle
	}
}

func supportedImageExtSet() map[string]struct{} {
	_, allow := imageFileFilters()
	out := make(map[string]struct{}, len(allow))
	for _, ext := range allow {
		out[strings.ToLower(strings.TrimSpace(ext))] = struct{}{}
	}
	return out
}

func isSupportedImagePath(path string) bool {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if ext == "" {
		return false
	}
	_, ok := supportedImageExtSet()[ext]
	return ok
}

func listSupportedImageFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !isSupportedImagePath(name) {
			continue
		}
		files = append(files, filepath.Join(dir, name))
	}
	sort.Strings(files)
	return files, nil
}

func pickRandomImagePath(candidates []string, last string, randIntn func(int) int) string {
	if len(candidates) == 0 {
		return ""
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	if randIntn == nil {
		randIntn = func(int) int { return 0 }
	}
	idx := randIntn(len(candidates))
	if idx < 0 || idx >= len(candidates) {
		idx = 0
	}
	picked := candidates[idx]
	if picked == last {
		return candidates[(idx+1)%len(candidates)]
	}
	return picked
}

func (f *FloatingWindow) ImageSourceMode() string {
	if f == nil {
		return imageSourceModeSingle
	}
	f.imageSourceMu.Lock()
	defer f.imageSourceMu.Unlock()
	return normalizeImageSourceMode(f.imageSourceMode)
}

func (f *FloatingWindow) FixedImagePath() string {
	if f == nil {
		return ""
	}
	f.imageSourceMu.Lock()
	defer f.imageSourceMu.Unlock()
	return strings.TrimSpace(f.fixedImagePath)
}

func (f *FloatingWindow) RandomFolderPath() string {
	if f == nil {
		return ""
	}
	f.imageSourceMu.Lock()
	defer f.imageSourceMu.Unlock()
	return strings.TrimSpace(f.randomFolderPath)
}

func (f *FloatingWindow) restoreImageSource() {
	if f == nil {
		return
	}

	mode := imageSourceModeSingle
	fixedPath := ""
	folderPath := ""
	if f.App != nil {
		prefs := f.App.Preferences()
		mode = normalizeImageSourceMode(prefs.String(imageSourceModeKey))
		fixedPath = strings.TrimSpace(prefs.String(fixedImagePathPrefKey))
		folderPath = strings.TrimSpace(prefs.String(randomFolderPathPrefKey))
		if fixedPath == "" {
			fixedPath = strings.TrimSpace(prefs.String("player.last_image_path"))
		}
	}
	if mode == imageSourceModeFolder && folderPath == "" {
		mode = imageSourceModeSingle
	}

	f.imageSourceMu.Lock()
	f.imageSourceMode = mode
	f.fixedImagePath = fixedPath
	f.randomFolderPath = folderPath
	f.lastRandomImagePath = ""
	f.imageSourceMu.Unlock()
}

func (f *FloatingWindow) saveImageSourceLocked() {
	if f == nil || f.App == nil {
		return
	}
	prefs := f.App.Preferences()
	prefs.SetString(imageSourceModeKey, normalizeImageSourceMode(f.imageSourceMode))
	prefs.SetString(fixedImagePathPrefKey, strings.TrimSpace(f.fixedImagePath))
	prefs.SetString(randomFolderPathPrefKey, strings.TrimSpace(f.randomFolderPath))
}

func (f *FloatingWindow) playImageOnStartup() {
	if f == nil {
		return
	}

	f.imageSourceMu.Lock()
	mode := normalizeImageSourceMode(f.imageSourceMode)
	fixedPath := strings.TrimSpace(f.fixedImagePath)
	folderPath := strings.TrimSpace(f.randomFolderPath)
	f.imageSourceMu.Unlock()

	if mode == imageSourceModeFolder && folderPath != "" {
		if f.playRandomFromFolder(folderPath) {
			f.startImageTicker()
			return
		}
	}

	f.stopImageTicker()
	if fixedPath != "" {
		if _, err := os.Stat(fixedPath); err == nil {
			f.playImagePath(fixedPath)
			return
		}
	}
	if f.Player != nil {
		f.Player.PlayLast()
	}
}

func (f *FloatingWindow) SetImageSourceMode(mode string) bool {
	if f == nil {
		return false
	}
	mode = normalizeImageSourceMode(mode)

	f.imageSourceMu.Lock()
	oldMode := normalizeImageSourceMode(f.imageSourceMode)
	folderPath := strings.TrimSpace(f.randomFolderPath)
	fixedPath := strings.TrimSpace(f.fixedImagePath)
	if mode == imageSourceModeFolder && folderPath == "" {
		f.imageSourceMu.Unlock()
		return false
	}
	f.imageSourceMode = mode
	f.saveImageSourceLocked()
	f.imageSourceMu.Unlock()

	if mode == imageSourceModeFolder {
		if !f.playRandomFromFolder(folderPath) {
			f.imageSourceMu.Lock()
			f.imageSourceMode = oldMode
			f.saveImageSourceLocked()
			f.imageSourceMu.Unlock()
			return false
		}
		f.startImageTicker()
		return true
	}

	f.stopImageTicker()
	if fixedPath != "" {
		f.playImagePath(fixedPath)
	}
	return true
}

func (f *FloatingWindow) SetFixedImage(path string) bool {
	if f == nil {
		return false
	}
	path = strings.TrimSpace(path)
	if path == "" || !isSupportedImagePath(path) {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}

	f.imageSourceMu.Lock()
	f.fixedImagePath = path
	f.imageSourceMode = imageSourceModeSingle
	f.saveImageSourceLocked()
	f.imageSourceMu.Unlock()

	f.stopImageTicker()
	f.playImagePath(path)
	return true
}

func (f *FloatingWindow) SetRandomImageFolder(dir string) bool {
	if f == nil {
		return false
	}
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return false
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	candidates, err := listSupportedImageFiles(dir)
	if err != nil || len(candidates) == 0 {
		return false
	}

	f.imageSourceMu.Lock()
	f.randomFolderPath = dir
	f.imageSourceMode = imageSourceModeFolder
	f.saveImageSourceLocked()
	f.imageSourceMu.Unlock()

	if !f.playRandomFromFolder(dir) {
		return false
	}
	f.startImageTicker()
	return true
}

func (f *FloatingWindow) PlayRandomImageNow() bool {
	if f == nil {
		return false
	}
	f.imageSourceMu.Lock()
	mode := normalizeImageSourceMode(f.imageSourceMode)
	dir := strings.TrimSpace(f.randomFolderPath)
	f.imageSourceMu.Unlock()
	if dir == "" {
		return false
	}
	ok := f.playRandomFromFolder(dir)
	if ok && mode == imageSourceModeFolder {
		f.startImageTicker()
	}
	return ok
}

func (f *FloatingWindow) playImagePath(path string) {
	if f == nil || f.Player == nil {
		return
	}
	f.Player.Play(path)
}

func (f *FloatingWindow) playRandomFromFolder(dir string) bool {
	candidates, err := listSupportedImageFiles(dir)
	if err != nil || len(candidates) == 0 {
		return false
	}

	f.imageSourceMu.Lock()
	last := f.lastRandomImagePath
	randIntn := f.randomIntn
	f.imageSourceMu.Unlock()

	picked := pickRandomImagePath(candidates, last, randIntn)
	if picked == "" {
		return false
	}
	f.playImagePath(picked)

	f.imageSourceMu.Lock()
	f.lastRandomImagePath = picked
	f.imageSourceMu.Unlock()
	return true
}

func (f *FloatingWindow) startImageTicker() {
	if f == nil {
		return
	}

	f.imageSourceMu.Lock()
	if f.imageTickerStop != nil {
		f.imageSourceMu.Unlock()
		return
	}
	interval := f.imageTickerInterval
	if interval <= 0 {
		interval = imageTickInterval
	}
	stop := make(chan struct{})
	f.imageTickerStop = stop
	f.imageSourceMu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				f.imageSourceMu.Lock()
				mode := normalizeImageSourceMode(f.imageSourceMode)
				dir := strings.TrimSpace(f.randomFolderPath)
				f.imageSourceMu.Unlock()
				if mode == imageSourceModeFolder && dir != "" {
					_ = f.playRandomFromFolder(dir)
				}
			case <-stop:
				return
			}
		}
	}()
}

func (f *FloatingWindow) stopImageTicker() {
	if f == nil {
		return
	}

	f.imageSourceMu.Lock()
	stop := f.imageTickerStop
	f.imageTickerStop = nil
	f.imageSourceMu.Unlock()
	if stop != nil {
		close(stop)
	}
}
