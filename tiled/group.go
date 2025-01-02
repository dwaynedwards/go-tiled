package tiled

type Groups []*Group

// WithName retrieves the first Group matching the provided name. Returns `nil` if not found.
func (gl Groups) WithName(name string) *Group {
	for _, g := range gl {
		if g.Name == name {
			return g
		}
	}
	return nil
}

type Group struct {
	Id        string  `xml:"id,attr"`
	Name      string  `xml:"name,attr"`
	Class     string  `xml:"class,attr"`
	Opacity   float32 `xml:"opacity,attr"`
	Visible   bool    `xml:"visible,attr"`
	OffsetX   int     `xml:"offsetx,attr"`
	OffsetY   int     `xml:"offsety,attr"`
	ParallaxX int     `xml:"parallaxx,attr"`
	ParallaxY int     `xml:"parallaxy,attr"`
	TintColor string  `xml:"tintcolor,attr"`

	Properties   *Properties   `xml:"properties>property"`
	TileLayers   *TileLayers   `xml:"layer"`
	ObjectLayers *ObjectLayers `xml:"objectgroup"`
	ImageLayers  *ImageLayers  `xml:"imagelayer"`
	Groups       *Groups       `xml:"group"`
}
