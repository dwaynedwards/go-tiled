package tiled_test

import (
	"fmt"
	"github.com/dwaynedwards/go-tiled/tiled"
	"github.com/matryer/is"
	"path/filepath"
	"runtime"
	"testing"
	"unsafe"
)

func TestMapLoading(t *testing.T) {
	is := is.New(t)

	parseFiles := []string{
		"../testdata/csv.tmx",
		"../testdata/b64zlib.tmx",
		"../testdata/b64zstd.tmx",
		"../testdata/externaltileset.tmx",
		"../testdata/objecttemplates.tmx",
	}

	for _, path := range parseFiles {
		t.Run(fmt.Sprintf("should parse file %s", filepath.Base(path)), func(t *testing.T) {
			var m1, m2 runtime.MemStats
			runtime.ReadMemStats(&m1)

			m, err := tiled.New(path)

			runtime.ReadMemStats(&m2)
			memoryUsage(m, &m1, &m2)

			is.NoErr(err) // Error parsing Map

			is.True(m.Properties.WithName("multilines").InnerValue == "foo\nbar\nbaz") // Property named `multilines` inner value should be `foo\nbar\nbaz`
			falseVal, err := m.Properties.WithName("bool_false").Bool()
			is.NoErr(err)      // Property named `bool_false` should be a bool
			is.True(!falseVal) // Property named `bool_false` value should be `false`
			trueVal, err := m.Properties.WithName("bool_true").Bool()
			is.NoErr(err)    // Property named `bool_true` should be a bool
			is.True(trueVal) // Property named `bool_true` value should be `true`

			ts := m.Tilesets.WithName("base")
			is.True(ts != nil)                            // Should have a Tileset named `base`
			is.Equal(ts.FirstGlobalID, tiled.GlobalID(1)) // Tileset FirstGlobalID should be `1`
			is.True(ts.Tiles.WithID(6).HasAnimation())    // Tileset tile 6 should have Animation
			is.True(ts.Image.Source == "numbers.png")

			g := m.Groups.WithName("Group")
			is.True(g != nil) // Should have a Group name `Group`

			il := g.ImageLayers.WithName("Image")
			is.True(il != nil)                   // Should have an Image layer name `Image`
			is.True(il.Image.Source == "bg.jpg") // Image layer Image source should be `bg.jpg`

			tl := g.TileLayers.WithName("Layer")
			is.True(tl != nil)                             // Should have a tile layer named `Layer`
			is.True(tl.TintColor == "#000000")             // Tile layer tint color should be `#000000`
			is.True(tl.RawData != nil)                     // Tile layer data should not be nil
			is.Equal(len(tl.TileDefs), tl.Width*tl.Height) // Tile layer tile defs and tile count should be equal

			td, err := tl.GetTileDefAtPosition(tl.Height-1, tl.Width-1)
			is.NoErr(err)      // Position should be in bounds
			is.True(td != nil) // Should get tile def
			td, err = tl.GetTileDefAtPosition(tl.Height, tl.Width)
			is.True(err != nil) // Position should be out of bounds
			is.True(td == nil)  // Should get no tile def
			td, err = tl.GetTileDefAtIndex((tl.Height * tl.Width) - 1)
			is.NoErr(err)      // Position should be in bounds
			is.True(td != nil) // Should get tile def
			td, err = tl.GetTileDefAtIndex(tl.Height * tl.Width)
			is.True(err != nil) // Position should be out of bounds
			is.True(td == nil)  // Should get no tile def

			ol := m.ObjectLayers.WithName("Objects")
			is.True(ol != nil) // Should have an Object layer name `Objects`
			is.Equal(ol.ParallaxX, float32(.12))
			is.Equal(ol.ParallaxY, float32(.12))
			is.True(ol.Objects.WithName("text").Text.Value == "Hello World") // Object with name `text` should have a value of `Hello World`
			is.True(ol.Objects.WithName("ellipse").IsEllipse())              // Object with name `ellipse` should be an ellipse
			is.True(ol.Objects.WithName("polygon").IsPolygon())              // Object with name `polygon` should be a polygon
			is.True(ol.Objects.WithName("polyline").IsPolyline())            // Object with name `polyline` should be a polyline
			is.True(ol.Objects.WithName("point").IsPoint())                  // Object with name `point` should be a point
		})
	}

}

func memoryUsage(m *tiled.Map, m1, m2 *runtime.MemStats) {
	fmt.Printf("Sizeof Map: %d\n", unsafe.Sizeof(*m))
	fmt.Println("Alloc:", m2.Alloc-m1.Alloc,
		"TotalAlloc:", m2.TotalAlloc-m1.TotalAlloc,
		"HeapAlloc:", m2.HeapAlloc-m1.HeapAlloc)
}
