package processer

import (
	"echosphere/apis"
	"echosphere/llm"
	"errors"
	"fmt"
	"time"
)

type Product struct {
	ID          int // ImtId
	TypeID      int // NmId
	Name        string
	Description string
	VendorCode  string
	Photos      []string
}

type ReviewProcessed struct {
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

type ReviewProcesser interface {
	ProcessReviews(reviews []apis.ReviewCondensed) ([]ReviewProcessed, error)
}

type LLMReviewProcesser struct {
	languageModel     *LanguageModel
	responseGenerator *ResponseGenerator
	moodRecognizer    *MoodRecognizer
	keywordFinder     *KeywordFinder
}

type LanguageModel struct {
	Llm     *llm.LargeLanguageModelAPI
	Context string
}

func NewLLMReviewProcesser(language_model llm.LargeLanguageModelAPI, responseGenerator ResponseGenerator, moodRecognizer MoodRecognizer, keywordFinder KeywordFinder) LLMReviewProcesser {
	return LLMReviewProcesser{
		languageModel: &LanguageModel{
			Llm:     &language_model,
			Context: "",
		},
		responseGenerator: &responseGenerator,
		moodRecognizer:    &moodRecognizer,
		keywordFinder:     &keywordFinder,
	}
}

func (p LLMReviewProcesser) ProcessReviews(reviews []apis.ReviewCondensed) ([]ReviewProcessed, error) {
	var res []ReviewProcessed
	for _, reviewUnprocessed := range reviews {
		processed := ReviewProcessed{
			ID:          reviewUnprocessed.ID,
			PublishedAt: reviewUnprocessed.PublishedAt,
			Status:      reviewUnprocessed.Status,
			Rating:      reviewUnprocessed.Rating,
			Text:        reviewUnprocessed.Text,
			Product:     Product(reviewUnprocessed.Product),
			Response:    "",
			Mood:        "",
			KeyWords:    "",
		}

		var err error
		processed.Mood, err = (*p.moodRecognizer).GetReviewMood(reviewUnprocessed, p.languageModel)
		if err != nil {
			return res, fmt.Errorf("Could not get mood of the review: %v", err)
		}

		// p.languageModel.Context = fmt.Sprintf("Учитывай, что у отзыва следующее настроение: %s. ", processed.Mood)
		processed.Response, err = (*p.responseGenerator).GenerateResponse(reviewUnprocessed, p.languageModel)
		if err != nil {
			return res, fmt.Errorf("Could not generate response to the review: %v", err)
		}

		// there is no need in context for finding key words
		// p.languageModel.Context = ""
		processed.KeyWords, err = (*p.keywordFinder).FindKeywords(reviewUnprocessed, p.languageModel)
		if err != nil {
			return res, fmt.Errorf("Could not find key words of the review: %v", err)
		}

		res = append(res, processed)
	}

	return res, nil
}

type LLMResponseGenerator struct{}

func (g LLMResponseGenerator) GenerateResponse(review apis.ReviewCondensed, language_model *LanguageModel) (string, error) {
	if language_model == nil {
		return "", errors.New("LLMResponseGenerator requires that language_model is not nil")
	}

	if review.Text != "" {
		reviewResponse, err := (*language_model.Llm).GenerateResponse(review, language_model.Context)
		if err != nil {
			return "", err
		}
		return reviewResponse, nil
	} else {
		// possibly generate some generic responses based on rating of the review
		return "", nil
	}
}

type LLMMoodRecognizer struct{}

func (r LLMMoodRecognizer) GetReviewMood(review apis.ReviewCondensed, language_model *LanguageModel) (string, error) {
	if language_model == nil {
		return "", errors.New("LLMMoodRecognizer requires that language_model is not nil")
	}

	if review.Text != "" {
		reviewMood, err := (*language_model.Llm).GetMood(review, language_model.Context)
		if err != nil {
			return "", err
		}
		return reviewMood, nil
	} else {
		return "", nil
	}
}

type LLMKeywordFinder struct{}

func (f LLMKeywordFinder) FindKeywords(review apis.ReviewCondensed, language_model *LanguageModel) (string, error) {
	if language_model == nil {
		return "", errors.New("LLMKeywordFinder requires that language_model is not nil")
	}

	if review.Text != "" {
		reviewKeywords, err := (*language_model.Llm).FindKeywords(review, language_model.Context)
		if err != nil {
			return "", err
		}
		return reviewKeywords, nil
	} else {
		return "", nil
	}
}
