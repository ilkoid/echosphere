package mock

import (
	"echosphere/apis"
	"echosphere/server"
	"fmt"
)

type MockMarketplace struct {
}

func (m *MockMarketplace) PostResponses(responses []server.MarketplaceResponse) error {
	for _, response := range responses {
		fmt.Println(response)
	}
	return nil
}

func (m *MockMarketplace) GetFeedback(config apis.FeedbackRequestConfig) ([]apis.ReviewCondensed, error) {
	return []apis.ReviewCondensed{}, nil
}
