package imageutil

import (
	"io"
	"image/jpeg"
	"image/png"
)

// convertJPEGToPNG converts from JPEG to PNG.
func ConvertJPEGToPNG(w io.Writer, r io.Reader) error {
	img, err := jpeg.Decode(r)
	if err != nil {
		return err
	}
	return png.Encode(w, img)
}
