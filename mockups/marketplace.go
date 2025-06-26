package mock

import (
	"echosphere/apis"
	server_structs "echosphere/server/structs"
	"fmt"
)

type MockMarketplace struct {
}

func (m *MockMarketplace) PostResponses(responses []server_structs.MarketplaceResponse) error {
	for _, response := range responses {
		fmt.Println(response)
	}
	return nil
}

func (m *MockMarketplace) GetFeedback(config apis.FeedbackRequestConfig) ([]apis.ReviewCondensed, error) {
	return []apis.ReviewCondensed{}, nil
}
