package tiled

import (
	"fmt"
	"strconv"
	"strings"
)

// Properties is an array of Property Objects
type Properties []*Property

// WithName retrieves the first Property with a given name, nil if none
func (pl Properties) WithName(name string) *Property {
	for _, p := range pl {
		if p.Name == name {
			return p
		}
	}
	return nil
}

// Property wraps any number of custom Properties, and is used as a child of a
// number of other Objects.
type Property struct {
	Name       string       `xml:"name,attr"`
	Type       PropertyType `xml:"type,attr"`
	CustomType string       `xml:"propertytype,attr"`
	Value      string       `xml:"value,attr"`
	InnerValue string       `xml:",chardata"`

	Properties *Properties `xml:"properties>property"`
}

// Float returns a value from a given float Property
func (p Property) Float() (v float64, err error) {
	if p.Type != Float {
		return v, fmt.Errorf("%w: float", ErrPropertyWrongType)
	}

	if v, err = strconv.ParseFloat(p.Value, 64); err != nil {
		return v, fmt.Errorf("%w: %w", ErrPropertyFailedConversion, err)
	}

	return
}

// Int returns a value from a given integer Property
func (p Property) Int() (v int64, err error) {
	if p.Type != Int {
		return v, fmt.Errorf("%w: int", ErrPropertyWrongType)
	}

	if v, err = strconv.ParseInt(p.Value, 10, 64); err != nil {
		return v, fmt.Errorf("%w: %w", ErrPropertyFailedConversion, err)
	}

	return
}

// Bool returns a value from a given boolean Property
func (p Property) Bool() (v bool, err error) {
	if p.Type != Bool {
		return v, fmt.Errorf("%w: bool", ErrPropertyWrongType)
	}

	return p.Value == "true", nil
}

type PropertyType int

const (
	String PropertyType = iota
	Int
	Float
	Bool
	Color
	File
	Obj
	Class
)

func (r *PropertyType) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	switch strings.ToLower(s) {
	default:
		return fmt.Errorf("%w: %s", ErrUnknownPropertyType, s)
	case "string":
		*r = String
	case "int":
		*r = Int
	case "float":
		*r = Float
	case "bool":
		*r = Bool
	case "color":
		*r = Color
	case "file":
		*r = File
	case "object":
		*r = Obj
	case "class":
		*r = Class
	}
	return nil
}
