package imageutil

import (
	"errors"
	"github.com/disintegration/imaging"
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

func GetTextFromMeme(img image.Image) (topText, botText string, err error) {
	b := img.Bounds()
	var texts [2]string
	rects := [2]image.Rectangle{
		image.Rect(b.Min.X, b.Min.Y, b.Max.X, b.Max.Y/2),
		image.Rect(b.Min.X, b.Max.Y/2, b.Max.X, b.Max.Y),
	}
	for i, rect := range rects {
		cropped := imaging.Crop(img, rect)
		croppedImg := cropped.SubImage(cropped.Rect)
		if cimg, ok := croppedImg.(ImageSet); ok {
			cb := croppedImg.Bounds()
			for x := cb.Min.X; x < cb.Max.X; x++ {
				for y := cb.Min.Y; y < cb.Max.Y; y++ {
					oldPixel := croppedImg.At(x, y)
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
			if err = png.Encode(file, croppedImg); err != nil {
				return
			}
			file.Sync()
			text := gosseract.Must(gosseract.Params{
				Src:       file.Name(),
				Languages: "eng",
			})
			stringPieces := strings.Split(text, "\n")
			for j, piece := range stringPieces {
				if len(piece) == 0 || len(strings.TrimSpace(piece)) == 0 {
					continue
				}
				if j != 0 {
					texts[i] += "\n"
				}
				texts[i] += piece
			}
		} else {
			err = errors.New("Image type does not support set method")
			return
		}
	}
	topText = texts[0]
	botText = texts[1]
	return
}
