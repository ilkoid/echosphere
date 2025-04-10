package server

import (
	wb "echosphere/apis/wb"
	"echosphere/apis"
	"echosphere/repository/domain"
	"echosphere/google_sheets"

	"fmt"
	"net/http"
)

const (
	spreadsheetID   = "1TKuEZvJsnqfkolxXtQzDPPUxdtDEVrtgjvCrVhnAA2o"
	sheetRange      = "Лист1!A1:B"
	credentialsFile = "/etc/google_sheets_credentials.json"
)

func addRoutes(
	mux *http.ServeMux,
	processer ReviewProcesser,
	repository Repository,
) {
	mux.Handle("/wbfeedback", wbHandler(repository, processer))
}

func wbHandler(repository Repository, processer ReviewProcesser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wbApi := wb.WBAPI{}
		config := apis.FeedbackRequestConfig{
			IsAnswered: false,
			Take:       10,
			Skip:       0,
			DateFrom:   nil,
			DateTo:     nil,
		}

		wb_reviews, _ := wbApi.GetFeedback(config)
		processedReviews, err := processer.ProcessReviews(wb_reviews)
		if err != nil {
			fmt.Printf("Error was: %v\n", err)
		}

		for _, review := range processedReviews {
			var photos []domain.Photo
			for _, photo := range review.Product.Photos {
				photos = append(photos, domain.Photo{ByteSlice: photo})
			}
			err := repository.AddProduct(
				domain.Product{
					VendorId:    review.Product.VendorCode,
					WBId:        review.Product.ID,
					Name:        review.Product.Name,
					Description: review.Product.Description,
				},
				photos...,
			)
			if err != nil {
				fmt.Printf("Couldnt add the product to the db: %v\n", err)
			}

			err = repository.AddReview(
				domain.Review{
					Id:                review.ID,
					PublishedAt:       review.PublishedAt,
					Rating:            review.Rating,
					Text:              review.Text,
					PublishedResponse: "",
					SuggestedResponse: review.Response,
					Mood:              review.Mood,
					KeyWords:          review.KeyWords,
				},
				domain.Product{
					VendorId:    review.Product.VendorCode,
					WBId:        review.Product.ID,
					Name:        review.Product.Name,
					Description: review.Product.Description,
				},
			)
			if err != nil {
				fmt.Printf("Couldnt add the review to the db: %v\n", err)
			}
		}

		values := google_sheets.PrepareDataForSheets(processedReviews)
		if err := google_sheets.WriteReviewsToSheet(values, spreadsheetID, sheetRange, credentialsFile); err != nil {
			http.Error(w, fmt.Sprintf("Failed to write to Google Sheets: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}
