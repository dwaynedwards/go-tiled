package tiled

import (
	"encoding/xml"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type byFirstGlobalID Tilesets

func (a byFirstGlobalID) Len() int           { return len(a) }
func (a byFirstGlobalID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byFirstGlobalID) Less(i, j int) bool { return a[i].FirstGlobalID < a[j].FirstGlobalID }

// Tilesets is an array of Tileset
type Tilesets []*Tileset

// WithName retrieves the first Tileset matching the provided name. Returns `nil` if not found.
func (tl Tilesets) WithName(name string) *Tileset {
	for _, t := range tl {
		if t.Name == name {
			return t
		}
	}
	return nil
}

// Tileset is a set of tiles, including the graphics data to be mapped to the tiles, and the actual arrangement of tiles.
type Tileset struct {
	FirstGlobalID   GlobalID        `xml:"firstgid,attr"`
	Source          string          `xml:"source,attr"`
	Name            string          `xml:"name,attr"`
	Class           string          `xml:"class,attr"`
	TileWidth       int             `xml:"tilewidth,attr"`
	TileHeight      int             `xml:"tileheight,attr"`
	Spacing         int             `xml:"spacing,attr"`
	Margin          int             `xml:"margin,attr"`
	TileCount       uint32          `xml:"tilecount,attr"`
	Columns         int             `xml:"columns,attr"`
	ObjectAlignment ObjectAlignment `xml:"objectalignment,attr"`

	Properties      *Properties      `xml:"properties>property"`
	TileOffset      *tileOffset      `xml:"tileOffset"`
	Image           *Image           `xml:"image"`
	TerrainTypes    *[]*Terrain      `xml:"terraintypes>terrain"`
	WangSets        *WangSets        `xml:"wangsets>wangset"`
	Tiles           *Tiles           `xml:"tile"`
	Transformations *Transformations `xml:"transformations"`
}

func (t *Tileset) HasImage() bool {
	return t.Image != nil
}

func (t *Tileset) HasTiles() bool {
	return t.Tiles != nil
}

func (t *Tileset) GetTileRect(tile *Tile) *Rect {
	return &Rect{
		Min: Point{int(tile.X), int(tile.Y)},
		Max: Point{int(tile.X + tile.Width), int(tile.Y + tile.Height)},
	}
}

func (t *Tileset) GetTileRectFromID(bareID uint32) *Rect {
	bID := int(bareID)
	fGID := int(t.FirstGlobalID)
	w := int(t.TileWidth)
	h := int(t.TileHeight)
	iw := int(t.Image.Width)
	ih := float64(t.Image.Height)
	count := int(math.Floor(float64(iw)/float64(w)) * math.Floor(ih/float64(h)))

	var tileHor = 0
	var tileVert = 0
	for i := 0; i < count; i++ {
		if i == bID-fGID {
			x := tileHor * w
			y := tileVert * h
			return &Rect{Point{x, y}, Point{x + w, y + h}}
		}

		// Update x and y position
		tileHor++

		if tileHor == iw/w {
			tileHor = 0
			tileVert++
		}
	}

	return nil
}

// Tiles is an array of Tile
type Tiles []*Tile

func (tl Tiles) WithID(id TileID) *Tile {
	for _, t := range tl {
		if t.TileID == id {
			return t
		}
	}
	return nil
}

// TileID is a tile id unique to each Tileset; often called the "local tile ID" in the Tiled docs.
type TileID uint32

// Tile represents an individual tile within a Tileset
type Tile struct {
	TileID      TileID  `xml:"id,attr"`
	X           int     `xml:"x,attr"`
	Y           int     `xml:"y,attr"`
	Width       int     `xml:"width,attr"`
	Height      int     `xml:"height,attr"`
	Probability float32 `xml:"probability,attr"`
	Type        string  `xml:"type,attr"`
	// Raw TerrainType loaded from XML. Not intended to be used directly; use (TerrainType). [Deprecated]
	RawTerrainType string `xml:"terrain,attr"`

	Properties  *Properties  `xml:"properties>property"`
	Image       *Image       `xml:"image"`
	Animation   *Animation   `xml:"animation>frame"`
	ObjectLayer *ObjectLayer `xml:"objectgroup"`

	TerrainType *TerrainType
}

func (t *Tile) HasImage() bool {
	return t.Image != nil
}

func (t *Tile) HasAnimation() bool {
	return t.Animation != nil
}

func (t *Tile) HasObjectLayer() bool {
	return t.ObjectLayer != nil
}

func (t *Tile) HasTerrainType() bool {
	return t.TerrainType != nil
}

// Terrain defines a type of terrain and its associated tile ID. [Deprecated]
type Terrain struct {
	Name       string      `xml:"name,attr"`
	TileID     TileID      `xml:"tile,attr"`
	Properties *Properties `xml:"properties>property"`
}

// TerrainType represents the unique corner tiles used by a particular terrain. [Deprecated]
type TerrainType struct {
	TopLeft     TileID
	TopRight    TileID
	BottomLeft  TileID
	BottomRight TileID
}

// Animation is an array for frame Objects
type Animation []*Frame

// Frame is a frame specifier in a given Animation
type Frame struct {
	TileID       TileID `xml:"tileid,attr"`
	DurationMsec int    `xml:"duration,attr"`
}

type Rect struct {
	Min Point
	Max Point
}

type tileOffset struct {
	X int `xml:"x,attr"`
	Y int `xml:"y,attr"`
}

// Transformations describes which transformations can be applied to the tiles (e.g. to extend a Wang set by
// transforming existing tiles).
type Transformations struct {
	// Whether the tiles in this set can be flipped horizontally (default 0)
	HFlip bool `xml:"hflip,attr"`
	// Whether the tiles in this set can be flipped vertically (default 0)
	VFlip bool `xml:"vflip,attr"`
	// Whether the tiles in this set can be rotated in 90 degree increments (default 0)
	Rotate bool `xml:"rotate,attr"`
	// Whether untransformed tiles remain preferred, otherwise transformed tiles are used to produce more variations
	// (default 0)
	PreferUntransformed bool `xml:"preferUntransformed,attr"`
}

// WangSets is an array of wangSet Objects
type WangSets []*WangSet

// WangSet Defines a list of colors and any number of Wang tiles using these colors.
type WangSet struct {
	Name   string `xml:"name,attr"`
	Class  string `xml:"class,attr"`
	TileID TileID `xml:"tile,attr"`

	Properties *Properties   `xml:"properties>property"`
	WangColors *[]*WangColor `xml:"wangcolor"`
	WangTiles  *[]*WangTile  `xml:"wangtile"`
}

// WangColor defines a color that can be used to define the corner and/or edge of a wangTile.
type WangColor struct {
	Name   string `xml:"name,attr"`
	Class  string `xml:"class,attr"`
	Color  string `xml:"color,attr"`
	TileID TileID `xml:"tile,attr"`

	Properties *Properties `xml:"properties>property"`
}

type WangID string

type WangTile struct {
	Name   string `xml:"name,attr"`
	TileID TileID `xml:"tileid,attr"`
	// WangID is a 32-bit unsigned integer stored in the format 0xCECECECE where C is a corner color and each E is an
	// edge color, from right to left clockwise, starting with the top edge.
	WangID WangID `xml:"wangid,attr"`
}

type ObjectAlignment int

const (
	Unspecified ObjectAlignment = iota
	TopLeft
	Top
	TopRight
	Left
	Center
	Right
	BottomLeft
	Bottom
	BottomRight
)

func (t *Tileset) UnmarshalXML(xd *xml.Decoder, start xml.StartElement) error {
	type tempTileSet Tileset
	var tmp tempTileSet

	if err := xd.DecodeElement(&tmp, &start); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTileset, err)
	}

	firstGlobalID := tmp.FirstGlobalID
	*t = (Tileset)(tmp)

	if tmp.Source == "" {
		return nil
	}

	path := filepath.Join(ResourcePath, tmp.Source)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open Tileset file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("error closing Tileset file handler %s", errors.Unwrap(err))
		}
	}(f)

	if err := xml.NewDecoder(f).Decode(&tmp); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTileset, err)
	}

	*t = (Tileset)(tmp)
	if firstGlobalID != 0 {
		t.FirstGlobalID = firstGlobalID
	}

	if t.HasImage() {
		return nil
	}

	var image *Image = nil

	if !t.HasTiles() {
		return fmt.Errorf("%w: tileset or tiles missing source image", ErrDecodingTileset)
	}

	for _, tile := range *t.Tiles {
		if !tile.HasImage() {
			continue
		}

		image = tile.Image
		break
	}

	if image == nil {
		return fmt.Errorf("%w: tileset or tiles missing source image", ErrDecodingTileset)
	}

	t.Image = image

	return nil
}

func (t *Tile) UnmarshalXML(xd *xml.Decoder, start xml.StartElement) error {
	type tempTile Tile
	var tmp tempTile

	if err := xd.DecodeElement(&tmp, &start); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTile, err)
	}

	*t = (Tile)(tmp)

	if t.RawTerrainType == "" {
		t.TerrainType = &TerrainType{}
		return nil
	}

	types := strings.Split(t.RawTerrainType, ",")

	if l := len(types); l != 4 {
		return fmt.Errorf(
			"unexpected terrain type specifier %v; expected 4 values, got %v",
			t.RawTerrainType,
			l,
		)
	}

	tid := make([]TileID, 4)
	for i := 0; i < len(types); i++ {
		id, err := strconv.ParseInt(strings.TrimSpace(types[i]), 10, 32)
		if err != nil {
			return err
		}
		tid[i] = TileID(id)
	}

	t.TerrainType = &TerrainType{
		TopLeft:     tid[0],
		TopRight:    tid[1],
		BottomLeft:  tid[2],
		BottomRight: tid[3],
	}

	return nil
}

func (o *ObjectAlignment) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownObjectAlignment, s)
	case "unspecified":
		*o = Unspecified
	case "topleft":
		*o = TopLeft
	case "top":
		*o = Top
	case "topright":
		*o = TopRight
	case "left":
		*o = Left
	case "center":
		*o = Center
	case "right":
		*o = Right
	case "bottomleft":
		*o = BottomLeft
	case "bottom":
		*o = Bottom
	case "bottomright":
		*o = BottomRight
	}
	return nil
}
