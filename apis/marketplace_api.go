package apis

import (
	"time"
)

type MarketplaceAPI interface {
	// GetFeedback() ([]Review, error)
	// GetNumberUnanswered() (NumberOfUnanswered, error)
}

type ReviewCondensed struct {
	ID          string
	PublishedAt time.Time // time.Time
	Status      string    // State
	Rating      int
	Text        string
	Product     Product // OzonItem,  WBProductDetails
}

type Product struct {
	ID     int // ImtId
	TypeID int // NmId
	Name   string
}

func ConvertOzonReview(ozonReview OzonReviewCondensed) ReviewCondensed {
	return ReviewCondensed{
		ID:          ozonReview.ID,
		PublishedAt: ozonReview.PublishedAt,
		Status:      ozonReview.Status,
		Rating:      int(ozonReview.Rating),
		Text:        ozonReview.Text,
		Product: Product{
			ID:     ozonReview.Product.ID,
			TypeID: ozonReview.Product.TypeID,
			Name:   ozonReview.Product.Name,
		},
	}
}

func ConvertWBReview(wbReview WBReviewCondensed) ReviewCondensed {
	return ReviewCondensed{
		ID:          wbReview.ID,
		PublishedAt: wbReview.CreatedDate,
		Status:      wbReview.State,
		Rating:      wbReview.ProductValuation,
		Text:        wbReview.Text,
		Product: Product{
			ID:     wbReview.Product.ProductDetails.ImtId, // ImtId как ID товара
			TypeID: wbReview.Product.ProductDetails.NmId,  // NmId как TypeID
			Name:   wbReview.Product.ProductDetails.ProductName,
		},
	}
}
