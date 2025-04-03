package processer

import (
	"echosphere/apis"
)

type MoodRecognizer interface {
	// language model is optional
	GetReviewMood(review apis.ReviewCondensed, language_model *LanguageModel) (string, error)
}
