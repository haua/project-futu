package player

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"golang.org/x/image/webp"
)

func PlayWebP(p *Player, path string, playbackID uint64) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("open webp failed: %v", err)
		return
	}
	defer f.Close()

	img, err := webp.Decode(f)
	if err != nil {
		log.Printf("decode webp failed: %v", err)
		return
	}
	if !p.isPlaybackActive(playbackID) {
		return
	}

	fyne.Do(func() {
		if !p.isPlaybackActive(playbackID) {
			return
		}
		b := img.Bounds()
		p.updateBaseSize(b.Dx(), b.Dy())
		p.Canvas.Image = img
		p.Canvas.Refresh()
	})
}
