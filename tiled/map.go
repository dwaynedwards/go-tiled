package tiled

import (
	"encoding/xml"
	"fmt"
	"sort"
	"strings"
)

// Map Tiled map definition  https://doc.mapeditor.org/en/stable/reference/tmx-map-format/
type Map struct {
	Version         string      `xml:"version,attr"`
	TiledVersion    string      `xml:"tiledversion,attr,omitempty"`
	Class           string      `xml:"class,attr"`
	Orientation     Orientation `xml:"orientation,attr"`
	RenderOrder     RenderOrder `xml:"renderorder,attr"`
	Width           int         `xml:"width,attr"`
	Height          int         `xml:"height,attr"`
	TileWidth       int         `xml:"tilewidth,attr"`
	TileHeight      int         `xml:"tileheight,attr"`
	HexSideLength   int         `xml:"hexsidelength,attr,omitempty"`
	StaggerAxis     string      `xml:"staggeraxis,attr,omitempty"`
	StaggerIndex    string      `xml:"staggerindex,attr,omitempty"`
	BackgroundColor string      `xml:"backgroundcolor,attr,omitempty"`
	NextLayerID     int         `xml:"nextlayerid,attr"`
	NextObjectID    int         `xml:"nextobjectid,attr"`
	Infinite        bool        `xml:"infinite,attr,omitempty"`

	Properties   *Properties   `xml:"properties>property"`
	Tilesets     *Tilesets     `xml:"tileset"`
	TileLayers   *TileLayers   `xml:"layer"`
	ObjectLayers *ObjectLayers `xml:"objectgroup"`
	ImageLayers  *ImageLayers  `xml:"imagelayer"`
	Groups       *Groups       `xml:"group"`
}

type Orientation int

const (
	Orthogonal Orientation = iota
	Isometric
	Staggered
	Hexagonal
)

type RenderOrder int

const (
	RightDown RenderOrder = iota
	RightUp
	LeftDown
	LeftUp
)

func (t *Map) UnmarshalXML(xd *xml.Decoder, start xml.StartElement) error {
	type tmpTilemap Map
	var tmp tmpTilemap

	if err := xd.DecodeElement(&tmp, &start); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTilemap, err)
	}

	*t = (Map)(tmp)

	sort.Sort(byFirstGlobalID(*t.Tilesets))

	if t.TileLayers != nil {
		for _, tl := range *t.TileLayers {
			if err := decodeTileDefs(tl, t.Tilesets); err != nil {
				return err
			}
		}
	}

	if err := decodeGroupTileDefs(t.Groups, t.Tilesets); err != nil {
		return err
	}

	return nil
}

func decodeGroupTileDefs(gl *Groups, tss *Tilesets) error {
	if gl == nil {
		return nil
	}

	for _, g := range *gl {
		if g.TileLayers != nil {
			for _, tl := range *g.TileLayers {
				if err := decodeTileDefs(tl, tss); err != nil {
					return err
				}
			}
		}

		if err := decodeGroupTileDefs(g.Groups, tss); err != nil {
			return err
		}
	}

	return nil
}

// TileDefs gets the definitions for all the tiles in a given TileLayer, matched with the given Tilesets
func decodeTileDefs(l *TileLayer, tss *Tilesets) (err error) {
	for _, tgr := range l.TileGlobalRefs {
		bid := tgr.GlobalID.BareID()

		if bid == 0 {
			l.TileDefs = append(l.TileDefs, &TileDef{Nil: true})
			continue
		}

		var ts *Tileset
		for _, i := range *tss {
			t := i
			if bid < uint32(t.FirstGlobalID) {
				break
			}

			ts = t
		}

		// if we never found a Tileset, the file is invalid; return an error that
		if ts == nil {
			return fmt.Errorf("%w, with global ID %v", ErrNoSuitableTileset, tgr.GlobalID)
		}

		var tile *Tile = nil
		id := tgr.GlobalID.TileID(ts)
		if ts.HasTiles() {
			tile = ts.Tiles.WithID(id)
		}
		l.TileDefs = append(l.TileDefs, &TileDef{
			ID:                  id,
			GlobalID:            tgr.GlobalID,
			TileSet:             ts,
			Tile:                tile,
			HorizontallyFlipped: tgr.GlobalID.IsFlippedHorizontally(),
			VerticallyFlipped:   tgr.GlobalID.IsFlippedVertically(),
			DiagonallyFlipped:   tgr.GlobalID.IsFlippedDiagonally(),
		})
	}
	// Release memory
	l.TileGlobalRefs = nil
	return nil
}

func (o *Orientation) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownOrientation, s)
	case "orthogonal":
		*o = Orthogonal
	case "isometric":
		*o = Isometric
	case "staggered":
		*o = Staggered
	case "hexagonal":
		*o = Hexagonal
	}
	return nil
}

func (r *RenderOrder) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownRenderOrder, s)
	case "right-down":
		*r = RightDown
	case "right-up":
		*r = RightUp
	case "left-down":
		*r = LeftDown
	case "left-up":
		*r = LeftUp
	}
	return nil
}
