package tiled

import (
	"fmt"
	"strings"
)

// Image represents a graphic asset to be used for a Tileset (or other element). While maps created with the Tiled
// editor may not have the Image embedded, the format can support it; no additional decoding or loading is attempted by
// this library, but the data will be available in the struct.
type Image struct {
	Format           ImageFormat `xml:"format,attr"`
	Source           string      `xml:"source,attr"`
	TransparentColor string      `xml:"trans,attr"`
	Width            int         `xml:"width,attr"`
	Height           int         `xml:"height,attr"`
	Data             *Data       `xml:"data"`
}

type ImageFormat int

const (
	Png ImageFormat = iota
	Gif
	Jpg
	Bmp
)

func (i *ImageFormat) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownImageFormat, s)
	case "png":
		*i = Png
	case "gif":
		*i = Gif
	case "jpg":
		*i = Jpg
	case "bmp":
		*i = Bmp
	}
	return nil
}
