package tiled

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zlib"
	"github.com/klauspost/compress/zstd"
)

// TileLayers is an array of TileLayer
type TileLayers []*TileLayer

// WithName retrieves the first TileLayer matching the provided name. Returns `nil` if not found.
func (tl TileLayers) WithName(name string) *TileLayer {
	for _, t := range tl {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// TileLayer aka <layer> specifies a TileLayer of a given Map; a TileLayer contains tile arrangement
// information.
type TileLayer struct {
	ID        string  `xml:"id,attr"`
	Name      string  `xml:"name,attr"`
	Class     string  `xml:"class,attr"`
	X         float32 `xml:"x,attr"`
	Y         float32 `xml:"y,attr"`
	Width     int     `xml:"width,attr"`
	Height    int     `xml:"height,attr"`
	Opacity   float32 `xml:"opacity,attr"`
	Visible   bool    `xml:"visible,attr"`
	TintColor string  `xml:"tintcolor,attr"`
	OffsetX   int     `xml:"offsetx,attr"`
	OffsetY   int     `xml:"offsety,attr"`
	ParallaxX int     `xml:"parallaxx,attr"`
	ParallaxY int     `xml:"parallaxy,attr"`

	Properties *Properties `xml:"properties>property"`
	// Raw data loaded from XML. Not intended to be used directly; use the TileGlobalRefs and TileDefs
	RawData *Data `xml:"data"`

	// Decoded data references
	TileGlobalRefs []*TileGlobalRef
	TileDefs       []*TileDef
}

func (l *TileLayer) GetTileDefAtPosition(row, col int) (*TileDef, error) {
	td, err := l.GetTileDefAtIndex((row * int(l.Width)) + col)
	if err != nil {
		return nil, fmt.Errorf("%w: row: %d, col: %d", ErrTileDefOutOfBounds, row, col)

	}
	return td, nil
}

func (l *TileLayer) GetTileDefAtIndex(index int) (*TileDef, error) {
	if index < 0 || index >= int(l.Width*l.Height) {
		return nil, fmt.Errorf("%w: index: %d", ErrTileDefOutOfBounds, index)
	}
	return l.TileDefs[index], nil
}

// Data represents a payload in a given Object; it may be specified in several different encodings and compressions, or as
// a straight data structure containing TileGlobalRefs
type Data struct {
	Encoding    string `xml:"encoding,attr"`
	Compression string `xml:"compression,attr"`
	// Raw data loaded from XML. Not intended to be used directly; use the layers TileGlobalRefs
	RawBytes []byte `xml:",innerxml"`
}

// TileGlobalRef is a reference to a tile GlobalID
type TileGlobalRef struct {
	GlobalID GlobalID `xml:"gid,attr"`
}

// TileDef is a representation of an individual hydrated tile, with all the necessary data to render that tile; it's
// built up off of the tile GlobalIDs, to give a TileLayer-local TileID, its Properties, and the Tileset used to render it
// (as a reference).
type TileDef struct {
	Nil                 bool
	ID                  TileID
	GlobalID            GlobalID
	TileSet             *Tileset
	Tile                *Tile
	HorizontallyFlipped bool
	VerticallyFlipped   bool
	DiagonallyFlipped   bool
}

// GlobalID is a per-map global unique ID used in TileLayer tile definitions (tileGlobalRef). It also encodes how the
// tile is drawn; if it's mirrored across an axis, for instance. Typically, you will not use a GlobalID directly; it
// will be mapped for you by various helper methods on other structs.
type GlobalID uint32

// IsFlippedHorizontally returns true if the ID specifies a horizontal flip
func (g GlobalID) IsFlippedHorizontally() bool {
	return g&TileFlippedHorizontally != 0
}

// IsFlippedVertically returns true if the ID specifies a vertical flip
func (g GlobalID) IsFlippedVertically() bool {
	return g&TileFlippedVertically != 0
}

// IsFlippedDiagonally returns true if the ID specifies a diagonal flip
func (g GlobalID) IsFlippedDiagonally() bool {
	return g&TileFlippedDiagonally != 0
}

// TileID returns the Tileset-relative TileID for a given GlobalID
func (g GlobalID) TileID(t *Tileset) TileID {
	return TileID(g.BareID() - uint32(t.FirstGlobalID))
}

// BareID returns the actual integer ID without tile flip information
func (g GlobalID) BareID() uint32 {
	return uint32(g &^ TileFlipped)
}

// Bitmasks for tile orientation
const (
	TileFlippedHorizontally = 0x80000000
	TileFlippedVertically   = 0x40000000
	TileFlippedDiagonally   = 0x20000000
	TileFlipped             = TileFlippedHorizontally | TileFlippedVertically | TileFlippedDiagonally
)

func (l *TileLayer) UnmarshalXML(xd *xml.Decoder, start xml.StartElement) error {
	type tempLayer TileLayer
	var tmp tempLayer

	if err := xd.DecodeElement(&tmp, &start); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTileLayer, err)
	}

	*l = (TileLayer)(tmp)

	if err := decodeLayerData(l); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTileLayerData, err)
	}

	return nil
}

func decodeLayerData(l *TileLayer) (err error) {
	switch l.RawData.Encoding {
	case "base64":
		b := base64.NewDecoder(base64.StdEncoding, bytes.NewReader(bytes.TrimSpace(l.RawData.RawBytes)))

		var r io.ReadCloser
		switch l.RawData.Compression {
		case "zlib":
			if r, err = zlib.NewReader(b); err != nil {
				return err
			}
		case "gzip":
			if r, err = gzip.NewReader(b); err != nil {
				return err
			}
		case "zstd":
			dd, err := zstd.NewReader(b)
			if err != nil {
				return err
			}
			defer dd.Close()
			dc, err := io.ReadAll(dd)
			if err != nil {
				return err
			}
			r = io.NopCloser(bytes.NewReader(dc))
		case "":
			r = io.NopCloser(b)
		default:
			return fmt.Errorf("%w: %s", ErrUnsupportedCompression, l.RawData.Compression)
		}
		defer func(r io.ReadCloser) {
			err := r.Close()
			if err != nil {
				fmt.Printf("failed to close decode layer data reader: %s", errors.Unwrap(err))
			}
		}(r)

		var nextInt uint32
		for {
			err := binary.Read(r, binary.LittleEndian, &nextInt)
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			l.TileGlobalRefs = append(l.TileGlobalRefs, &TileGlobalRef{
				GlobalID: GlobalID(nextInt),
			})
		}
	case "csv":
		for _, s := range strings.Split(string(l.RawData.RawBytes), ",") {
			nextInt, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
			if err != nil {
				return err
			}

			l.TileGlobalRefs = append(l.TileGlobalRefs, &TileGlobalRef{
				GlobalID: GlobalID(uint32(nextInt)),
			})
		}
	case "":
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedEncoding, l.RawData.Encoding)
	}

	return nil
}
