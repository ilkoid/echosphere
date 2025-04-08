package domain

import "time"

type Product struct {
	VendorId    int
	WBId        int
	Name        string
	Description string
}

type Photo struct {
	Id        int
	ByteSlice string
}

type ProductPhotos struct {
	ProductVendorCode int
	PhotoId           int
}

type Review struct {
	Id                int
	PublishedAt       time.Time
	Rating            int
	Text              string
	PublishedResponse string
	SuggestedResponse string
	Mood              string
	KeyWords          string
}

type ReviewOfProduct struct {
	ReviewId          int
	ProductVendorCode int
}
