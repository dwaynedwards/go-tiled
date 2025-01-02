package tiled

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ObjectLayers is an array of ObjectLayer
type ObjectLayers []*ObjectLayer

// WithName retrieves the first ObjectLayer matching the provided name. Returns `nil` if not found.
func (ol ObjectLayers) WithName(name string) *ObjectLayer {
	for _, o := range ol {
		if o.Name == name {
			return o
		}
	}
	return nil
}

// ObjectLayer aka <objectgroup> is a Group of Objects within a Map or tile, used to specify sub-Objects such as polygons.
type ObjectLayer struct {
	ID        string    `xml:"nid,attr"`
	Name      string    `xml:"name,attr"`
	Class     string    `xml:"class,attr"`
	Color     string    `xml:"color,attr"`
	X         float32   `xml:"x,attr"`
	Y         float32   `xml:"y,attr"`
	Width     int       `xml:"width,attr"`
	Height    int       `xml:"height,attr"`
	Opacity   float32   `xml:"opacity,attr"`
	Visible   bool      `xml:"visible,attr"`
	OffsetX   int       `xml:"offsetx,attr"`
	OffsetY   int       `xml:"offsety,attr"`
	ParallaxX float32   `xml:"parallaxx,attr"`
	ParallaxY float32   `xml:"parallaxy,attr"`
	DrawOrder DrawOrder `xml:"draworder,attr"`

	Properties *Properties `xml:"properties>property"`
	Objects    *Objects    `xml:"object"`
}

// Objects is an array of Object Objects
type Objects []*Object

// WithName retrieves the first Object with a given name, nil if none
func (ol Objects) WithName(name string) *Object {
	for _, o := range ol {
		if o.Name == name {
			return o
		}
	}

	return nil
}

// ObjectID specifies a unique ID
type ObjectID uint32

// Object is an individual Object, such as a Polygon, Polyline, or otherwise.
type Object struct {
	ObjectID ObjectID `xml:"id,attr"`
	Name     string   `xml:"name,attr"`
	Type     string   `xml:"type,attr"`
	X        float32  `xml:"x,attr"`
	Y        float32  `xml:"y,attr"`
	Width    float32  `xml:"width,attr"`
	Height   float32  `xml:"height,attr"`
	Rotation float32  `xml:"rotation,attr"`
	Visible  bool     `xml:"visible,attr"`
	Template string   `xml:"template,attr"`
	GlobalID GlobalID `xml:"gid,attr"`

	Properties *Properties `xml:"properties>property"`
	Image      *Image      `xml:"image"`
	Polygon    *Poly       `xml:"polygon"`
	Polyline   *Poly       `xml:"polyline"`
	Text       *Text       `xml:"text"`
	Point      *struct{}   `xml:"point"`
	Ellipse    *struct{}   `xml:"ellipse"`
}

// IsPoint returns true if the Object is a point, else false
func (o *Object) IsPoint() bool {
	return o.Point != nil
}

// IsPolygon returns true if the Object is a polygon, else false
func (o *Object) IsPolygon() bool {
	return o.Polygon != nil
}

// IsPolyline returns true if the Object is a polyline, else false
func (o *Object) IsPolyline() bool {
	return o.Polyline != nil
}

// IsEllipse returns true if the Object is an ellipse, else false
func (o *Object) IsEllipse() bool {
	return o.Ellipse != nil
}

// IsText returns true if the Object is text, else false
func (o *Object) IsText() bool {
	return o.Text != nil
}

type Text struct {
	FontFamily string     `xml:"fontfamily,attr"`
	PixelSize  int        `xml:"pixelsize,attr"`
	Wrap       bool       `xml:"wrap,attr"`
	Bold       bool       `xml:"bold,attr"`
	Italic     bool       `xml:"italic,attr"`
	Underline  bool       `xml:"underline,attr"`
	Strikeout  bool       `xml:"strikeout,attr"`
	Kerning    bool       `xml:"kerning,attr"`
	HAlign     HAlignment `xml:"halign,attr"`
	VAlign     VAlignment `xml:"valign,attr"`
	Value      string     `xml:",chardata"`
}

// Point is an X, Y coordinate in space
type Point struct {
	X, Y int
}

// Poly represents a collection of points; used to represent a Polyline or a Polygon
type Poly struct {
	// Raw Points loaded from XML. Not intended to be used directly; use the
	// methods on this struct to accessed parsed data.
	RawPoints string `xml:"points,attr"`
}

// Points returns a list of points in a Poly
func (p *Poly) Points() (pts []Point, err error) {
	for _, rpt := range strings.Split(p.RawPoints, " ") {
		var x, y int64

		xy := strings.Split(rpt, ",")
		if l := len(xy); l != 2 {
			err = fmt.Errorf(
				"unexpected number of coordinates in point destructure: %v in %v",
				l, rpt,
			)

			return
		}

		x, err = strconv.ParseInt(xy[0], 10, 32)
		if err != nil {
			return
		}
		y, err = strconv.ParseInt(xy[1], 10, 32)
		if err != nil {
			return
		}

		pts = append(pts, Point{int(x), int(y)})
	}
	return
}

type Template struct {
	TileSet *Tileset `xml:"tileset"`
	Object  *Object  `xml:"object"`
}

type DrawOrder int

const (
	TopDown DrawOrder = iota
	Index
)

type HAlignment int

const (
	HLeft HAlignment = iota
	HCenter
	HRight
	HJustify
)

type VAlignment int

const (
	VTop VAlignment = iota
	VCenter
	VBottom
)

func (t *ObjectLayer) UnmarshalXML(xd *xml.Decoder, start xml.StartElement) error {
	type tmpObjectLayer ObjectLayer
	var tmp tmpObjectLayer

	if err := xd.DecodeElement(&tmp, &start); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingObjectLayer, err)
	}

	*t = (ObjectLayer)(tmp)

	return nil
}

func (o *Object) UnmarshalXML(xd *xml.Decoder, start xml.StartElement) error {
	type tmpObject Object
	var tmp tmpObject

	if err := xd.DecodeElement(&tmp, &start); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTileLayer, err)
	}

	*o = (Object)(tmp)

	if tmp.Template == "" {
		return nil
	}

	path := filepath.Join(ResourcePath, tmp.Template)
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open template file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("error closing template file handler %s", errors.Unwrap(err))
		}
	}(f)

	var template Template
	if err := xml.NewDecoder(f).Decode(&template); err != nil {
		return fmt.Errorf("%w: %w", ErrDecodingTemplate, err)
	}

	if o.Name == "" {
		o.Name = tmp.Name
	}
	if o.Type == "" {
		o.Type = tmp.Type
	}
	if o.X == 0 {
		o.X = tmp.X
	}
	if o.Y == 0 {
		o.Y = tmp.Y
	}
	if o.Width == 0 {
		o.Width = tmp.Width
	}
	if o.Height == 0 {
		o.Height = tmp.Height
	}
	if o.Rotation == 0 {
		o.Rotation = tmp.Rotation
	}
	if !o.Visible {
		o.Visible = tmp.Visible
	}
	if o.GlobalID == 0 {
		o.GlobalID = tmp.GlobalID
	}
	if o.Properties == nil {
		o.Properties = tmp.Properties
	}
	if o.Image == nil {
		o.Image = tmp.Image
	}
	if o.Polygon == nil {
		o.Polygon = tmp.Polygon
	}
	if o.Polygon == nil {
		o.Polyline = tmp.Polyline
	}
	if o.Text == nil {
		o.Text = tmp.Text
	}
	if o.Ellipse == nil {
		o.Ellipse = tmp.Ellipse
	}
	if o.Point == nil {
		o.Point = tmp.Point
	}

	return nil
}

func (d *DrawOrder) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownDrawOrder, s)
	case "":
		*d = TopDown
	case "topdown":
		*d = TopDown
	case "index":
		*d = Index
	}
	return nil
}

func (o *HAlignment) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownHAlignment, s)
	case "":
		*o = HLeft
	case "left":
		*o = HLeft
	case "center":
		*o = HCenter
	case "right":
		*o = HRight
	case "justify":
		*o = HJustify
	}
	return nil
}

func (o *VAlignment) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownVAlignment, s)
	case "":
		*o = VTop
	case "top":
		*o = VTop
	case "center":
		*o = VCenter
	case "bottom":
		*o = VBottom
	}
	return nil
}
