package domain

import "time"

type Product struct {
	VendorId    string
	WBId        int
	Name        string
	Description string
}

type Photo struct {
	ByteSlice string
}

type ProductPhotos struct {
	ProductVendorCode string
	PhotoId           int
}

type Review struct {
	Id                string
	PublishedAt       time.Time
	Rating            int
	Text              string
	PublishedResponse string
	SuggestedResponse string
	Mood              string
	KeyWords          string
}

type ReviewOfProduct struct {
	ReviewId          string
	ProductVendorCode string
}
