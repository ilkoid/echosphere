package server

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	"strconv"

	"gopkg.in/yaml.v3"
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
	marketplace MarketplaceAPI,
) {
	mux.Handle("/api/v1/reviews/list", getReviewsHandler(repository))
	mux.Handle("/api/v1/reviews/publish", publishReviewsHandler(repository, marketplace))
}

func publishReviewsHandler(repository Repository, marketplace MarketplaceAPI) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "This endpoint is POST method", http.StatusBadRequest)
			return
		}

		updates, err := decode[[]MarketplaceResponse](r)
		if err != nil {
			http.Error(w, "UpdateReviews struct could not be unmarshalled", http.StatusBadRequest)
			return
		}

		// instead of this cycle will be post method to the WB
		marketplace.PostResponses(updates)

		err = repository.UpdateReviews(updates)
		if err != nil {
			http.Error(w, "Error writing to the repository", http.StatusInternalServerError)
			return

		}

		w.WriteHeader(http.StatusOK)
	})
}

func getReviewsHandler(repository Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "This endpoint is GET method", http.StatusBadRequest)
			return
		}

		type ReviewData struct {
			ID                string `json:"id"`
			Text              string `json:"text"`
			Rating            int    `json:"rating"`
			SuggestedResponse string `json:"suggested_response"`
			Mood              string `json:"mood"`
			KeyWords          string `json:"key_words"`
		}

		type ProductData struct {
			WbId        int      `json:"wb_id"`
			VendorId    string   `json:"vendor_id"`
			Name        string   `json:"name"`
			Description string   `json:"description"`
			Photos      []string `json:"photos"`
		}

		type Response struct {
			ReviewData  ReviewData  `json:"review"`
			ProductData ProductData `json:"product"`
		}

		dateFromStr := r.URL.Query().Get("dateFrom")
		if dateFromStr == "" {
			http.Error(w, "Date parameter required", http.StatusBadRequest)
			return
		}

		dateFrom, err := strconv.ParseInt(dateFromStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid date parameter", http.StatusBadRequest)
			return
		}

		filter := Filter{
			DateFrom: dateFrom,
			Rating:   nil,
			VendorId: "",
		}
		ratingStr := r.URL.Query().Get("rating")
		if ratingStr != "" {
			rating, err := strconv.Atoi(ratingStr)
			if err != nil {
				http.Error(w, "Invalid rating parameter", http.StatusBadRequest)
				return
			}
			filter.Rating = &rating
		}

		vendorId := r.URL.Query().Get("vendorId")
		if vendorId != "" {
			filter.VendorId = vendorId
		}

		cards, err := repository.GetCardsByFilter(filter)
		if err != nil {
			errStr := fmt.Sprintf("Failed to get reviews: %v", err)
			http.Error(w, errStr, http.StatusInternalServerError)
			return
		}

		var response []Response
		for _, card := range cards {
			var photos []string
			for _, photo := range card.Photos {
				photos = append(photos, photo.ByteSlice)
			}
			response = append(response, Response{
				ReviewData: ReviewData{
					ID:                card.Review.Id,
					Text:              card.Review.Text,
					Rating:            card.Review.Rating,
					SuggestedResponse: card.Review.SuggestedResponse,
					Mood:              card.Review.Mood,
					KeyWords:          card.Review.KeyWords,
				},
				ProductData: ProductData{
					WbId:        card.Product.WBId,
					VendorId:    card.Product.VendorId,
					Name:        card.Product.Name,
					Description: card.Product.Description,
					Photos:      photos,
				},
			})
		}

		w.Header().Set("Content-Type", "application/json")
		err = encode(w, int(http.StatusOK), response)
		if err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	})
}

func wbHandler(repository Repository, processer ReviewProcesser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type Config struct {
			IsAnswered bool `yaml:"is_answered"`
			Take       int  `yaml:"take"`
			Skip       int  `yaml:"skip"`
		}

		wbApi := wb.WBAPI{}

		file, _ := os.Open("config.yaml")
		defer file.Close()

		var conf Config
		bytes, _ := io.ReadAll(file)
		yaml.Unmarshal(bytes, &conf)

		config := apis.FeedbackRequestConfig{
			IsAnswered: conf.IsAnswered,
			Take:       conf.Take,
			Skip:       conf.Skip,
			DateFrom:   nil,
			DateTo:     nil,
		}

		wb_reviews, _ := wbApi.GetFeedback(config)
		processedReviews, err := processer.ProcessReviews(wb_reviews)
		if err != nil {
			fmt.Printf("Error was: %v\n", err)
		}

		for _, review := range processedReviews {
			product, photos := ConvertIntoDomainProduct(review)
			domainReview := ConvertIntoDomainReview(review)

			exists, err := repository.IsProductExist(product)
			if err != nil {
				fmt.Printf("Couldnt add the product to the db: %v\n", err)
			}
			if !exists {
				err = repository.AddProduct(product, photos...)
				if err != nil {
					fmt.Printf("Couldnt add the product to the db: %v\n", err)
				}
			}

			exists, err = repository.IsReviewExist(domainReview)
			if err != nil {
				fmt.Printf("Couldnt add the review to the db: %v\n", err)
			}
			if !exists {
				err = repository.AddReview(domainReview, product)
				if err != nil {
					fmt.Printf("Couldnt add the review to the db: %v\n", err)
				}
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
