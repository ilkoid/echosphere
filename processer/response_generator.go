package processer

import (
	"echosphere/apis"
)

type ResponseGenerator interface {
	// language model is optional
	GenerateResponse(review apis.ReviewCondensed, language_model *LanguageModel) (string, error)
}
