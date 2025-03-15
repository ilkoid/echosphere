package llm

import "echosphere/apis"

type LargeLanguageModelAPI interface {
	MakeResponse(reviews []apis.ReviewCondensed) ([]apis.ReviewCondensed, error)
	// GetFeelings(reviews []apis.ReviewCondensed) []apis.ReviewCondensed // to be added later
}
