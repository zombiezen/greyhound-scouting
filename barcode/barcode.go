// barcode.go

package barcode

import (
	"image"
	"image/color"
)

// Barcode is a 1-dimensional barcode.
type Barcode []bool

func (code Barcode) String() string {
	chars := make([]byte, len(code))
	for i, b := range code {
		if b {
			chars[i] = '1'
		} else {
			chars[i] = '0'
		}
	}
	return string(chars)
}

// Image renders a barcode through the image interface.
type Image struct {
	Barcode
	Scale  int
	Height int
}

func (img *Image) ColorModel() color.Model {
	return color.GrayModel
}

func (img *Image) Bounds() image.Rectangle {
	return image.Rect(0, 0, len(img.Barcode)*img.Scale, img.Height)
}

func (img *Image) At(x, y int) color.Color {
	if x < 0 || y < 0 || x >= len(img.Barcode)*img.Scale || y >= img.Height {
		return nil
	}

	if img.Barcode[x/img.Scale] {
		return color.Gray{0x00}
	}
	return color.Gray{0xff}
}
