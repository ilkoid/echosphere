package server

import (
	"echosphere/processer"
	"echosphere/repository/domain"
)

// implement conversion of data type from one package to another

func ConvertIntoDomainProduct(review processer.ReviewProcessed) (domain.Product, []domain.Photo) {
	var photos []domain.Photo
	for _, photo := range review.Product.Photos {
		photos = append(photos, domain.Photo{ByteSlice: photo})
	}
	return domain.Product{
		VendorId:    review.Product.VendorCode,
		WBId:        review.Product.ID,
		Name:        review.Product.Name,
		Description: review.Product.Description,
	}, photos
}

func ConvertIntoDomainReview(review processer.ReviewProcessed) domain.Review {
	return domain.Review{
		Id:                review.ID,
		PublishedAt:       review.PublishedAt,
		Rating:            review.Rating,
		Text:              review.Text,
		PublishedResponse: "",
		SuggestedResponse: review.Response,
		Mood:              review.Mood,
		KeyWords:          review.KeyWords,
	}
}
