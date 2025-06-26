package server_structs

import "echosphere/repository/domain"

type Card struct {
	Review  domain.Review
	Product domain.Product
	Photos  []domain.Photo
}

type Filter struct {
	DateFrom int64
	Rating   *int
	VendorId string
}

type MarketplaceResponse struct {
	Id   string `json:"id"`
	Text string `json:"text"`
}
