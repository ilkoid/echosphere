package apis

import "echosphere/apis"

func ConvertOzonReview(ozonReview OzonReviewCondensed) apis.ReviewCondensed {
	return apis.ReviewCondensed{
		ID:          ozonReview.ID,
		PublishedAt: ozonReview.PublishedAt,
		Status:      ozonReview.Status,
		Rating:      int(ozonReview.Rating),
		Text:        ozonReview.Text,
		Product: apis.Product{
			ID:     ozonReview.Product.ID,
			TypeID: ozonReview.Product.TypeID,
			Name:   ozonReview.Product.Name,
		},
	}
}
