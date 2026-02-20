package player

import (
	"log"
	"os"

	"fyne.io/fyne/v2"
	"golang.org/x/image/webp"
)

func PlayWebP(p *Player, path string) {
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

	fyne.Do(func() {
		p.Canvas.Image = img
		p.Canvas.Refresh()
	})
}
