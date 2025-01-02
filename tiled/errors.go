package tiled

import "errors"

// Possible Errors
var (
	ErrUnsupportedEncoding      = errors.New("invalid encoding")
	ErrUnsupportedCompression   = errors.New("unsupported compression type")
	ErrNoSuitableTileset        = errors.New("no suitable Tileset found for tile")
	ErrPropertyWrongType        = errors.New("a Property was found, but its type was incorrect")
	ErrPropertyFailedConversion = errors.New("the Property failed to convert to the expected type")
	ErrUnknownOrientation       = errors.New("unknown orientation type")
	ErrUnknownRenderOrder       = errors.New("unknown render order type")
	ErrUnknownObjectAlignment   = errors.New("unknown Object alignment type")
	ErrUnknownHAlignment        = errors.New("unknown horizontal alignment type")
	ErrUnknownVAlignment        = errors.New("unknown vertical alignment type")
	ErrUnknownImageFormat       = errors.New("unknown Image format type")
	ErrUnknownDrawOrder         = errors.New("unknown draw order type")
	ErrUnknownPropertyType      = errors.New("unknown Property type")
	ErrDecodingTilemap          = errors.New("failed to decode tilemap")
	ErrDecodingTileset          = errors.New("failed to decode tileset")
	ErrDecodingTile             = errors.New("failed to decode tile")
	ErrDecodingTileLayer        = errors.New("failed to decode tile layer")
	ErrDecodingTileLayerData    = errors.New("failed to decode tile layer data")
	ErrDecodingObjectLayer      = errors.New("failed to decode object layer")
	ErrDecodingTemplate         = errors.New("failed to decode template")
	ErrTileDefOutOfBounds       = errors.New("failed to get tile def out of bounds")
)
