package apis

import (
	"time"
)

type FeedbackRequestConfig struct {
	IsAnswered bool
	Take       int
	Skip       int
	DateFrom   *time.Time
	DateTo     *time.Time
}

type MarketplaceAPI interface {
	GetFeedback(config FeedbackRequestConfig) ([]ReviewCondensed, error)
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
	Response    string  // "" when there is no answer
	Mood        string  // "" when initialized
	KeyWords    string  // "" when initialized
}

type Product struct {
	ID     int // ImtId
	TypeID int // NmId
	Name   string
}
