package model

type Region struct {
	ID         string  `db:"id" json:"id"`
	RegionType string  `db:"region_type" json:"region_type"`
	IsoA2      string  `db:"iso_a2" json:"iso_a2"`
	Iso31662   *string  `db:"iso_3166_2" json:"iso_3166_2,omitempty"`
	PostalCode *string  `db:"postal" json:"postal,omitempty"`
	Name       string  `db:"name" json:"name"`
	Adm0a3     string  `db:"adm0_a3" json:"adm0_a3"`
	Latitude   *float64 `db:"latitude" json:"latitude,omitempty"`
	Longitude  *float64 `db:"longitude" json:"longitude,omitempty"`
}
