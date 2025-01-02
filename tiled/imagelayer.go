package tiled

type ImageLayers []*ImageLayer

// WithName retrieves the first ImageLayer matching the provided name. Returns `nil` if not found.
func (il ImageLayers) WithName(name string) *ImageLayer {
	for _, i := range il {
		if i.Name == name {
			return i
		}
	}
	return nil
}

// ImageLayer is a TileLayer consisting of a single Image, such as a background.
type ImageLayer struct {
	ID        string  `xml:"id,attr"`
	Name      string  `xml:"name,attr"`
	Class     string  `xml:"class,attr"`
	X         int     `xml:"x,attr"`
	Y         int     `xml:"y,attr"`
	OffsetX   int     `xml:"offsetx,attr"`
	OffsetY   int     `xml:"offsety,attr"`
	ParallaxX int     `xml:"parallaxx,attr"`
	ParallaxY int     `xml:"parallaxy,attr"`
	Opacity   float32 `xml:"opacity,attr"`
	Visible   bool    `xml:"visible,attr"`
	TintColor string  `xml:"tintcolor,attr"`
	RepeatX   bool    `xml:"repeatx,attr"`
	RepeatY   bool    `xml:"repeaty,attr"`

	Properties *Properties `xml:"properties>property"`
	Image      *Image      `xml:"image"`
}
