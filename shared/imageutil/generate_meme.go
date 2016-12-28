package imageutil

import (
	"github.com/harrisonzhao/supermeme/models"
	"github.com/nomad-software/meme/cli"
	"github.com/nomad-software/meme/image/renderer"
	"image"
)

func CreateMemeFromImage(meme models.Meme, img image.Image) image.Image {
	options := cli.Options{
		Bottom: meme.BottomText,
		Top:    meme.TopText,
	}
	return renderer.Render(options, img)
}
