package apis

import (
	"time"
)

type MarketplaceAPI interface {
	// GetFeedback() ([]Review, error)
	// GetNumberUnanswered() (NumberOfUnanswered, error)
}

// TODO: ADD BARCODE
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
