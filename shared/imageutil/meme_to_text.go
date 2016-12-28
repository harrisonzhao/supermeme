package imageutil

import (
	"errors"
	"github.com/otiai10/gosseract"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
)

type ImageSet interface {
	Set(x, y int, c color.Color)
}

// This function sometimes does not successfully separate top from bottom text
// Need to cut image in half and save each half separately to get top and bottom
func GetTextFromMeme(img image.Image) (topText, botText string, err error) {
	b := img.Bounds()
	topText = ""
	botText = ""
	if cimg, ok := img.(ImageSet); ok {
		for x := b.Min.X; x < b.Max.X; x++ {
			for y := b.Min.Y; y < b.Max.Y; y++ {
				oldPixel := img.At(x, y)
				if color.Gray16Model.Convert(oldPixel) != color.White {
					cimg.Set(x, y, color.Black)
				}
			}
		}
		var file *os.File
		file, err = ioutil.TempFile(os.TempDir(), "prefix")
		if err != nil {
			return
		}
		defer os.Remove(file.Name())
		if err = png.Encode(file, img); err != nil {
			return
		}
		file.Sync()
		out := gosseract.Must(gosseract.Params{
			Src:       file.Name(),
			Languages: "eng",
		})
		topText, botText = getTopBotText(out)
	} else {
		err = errors.New("Image type does not support set method")
	}
	return
}

func getTopBotText(txt string) (topText, botText string) {
	stringPieces := strings.Split(txt, "\n")
	isTop := true
	for i, piece := range stringPieces {
		if len(piece) == 0 || len(strings.TrimSpace(piece)) == 0 {
			isTop = false
			continue
		}
		str := piece
		if i != len(stringPieces)-1 {
			str += "\n"
		}
		if isTop {
			topText += str
		} else {
			botText += str
		}
	}
	return
}
