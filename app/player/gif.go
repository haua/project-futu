package player

import (
	"log"
	"image/gif"
	"os"
	"time"

	"fyne.io/fyne/v2"
)

func PlayGIF(p *Player, path string) {
	f, err := os.Open(path)
	if err != nil {
		log.Printf("open gif failed: %v", err)
		return
	}
	defer f.Close()

	g, err := gif.DecodeAll(f)
	if err != nil {
		log.Printf("decode gif failed: %v", err)
		return
	}

	go func() {
		for {
			for i, frame := range g.Image {
				delay := time.Duration(g.Delay[i]) * 10 * time.Millisecond
				fyne.Do(func() {
					p.Canvas.Image = frame
					p.Canvas.Refresh()
				})
				time.Sleep(delay)
			}
		}
	}()
}
