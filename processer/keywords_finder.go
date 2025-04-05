package processer

import (
	"echosphere/apis"
)

type KeywordFinder interface {
	// language model is optional
	// returns merged string of all key words in a review
	FindKeywords(review apis.ReviewCondensed, language_model *LanguageModel) (string, error)
}
