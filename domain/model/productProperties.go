package model

// ProductProperties represents the properties field of Product entity
type ProductProperties struct {
	// Size represents the size of the product (S, M, L)
	Size *string `json:"size,omitempty"`
	// Latitude represents the latitude coordinate
	Latitude *string `json:"latitude,omitempty"`
	// Longitude represents the longitude coordinate
	Longitude *string `json:"longitude,omitempty"`
	// Color represents the color of the product (red, green, blue)
	Color *string `json:"color,omitempty"`
}

