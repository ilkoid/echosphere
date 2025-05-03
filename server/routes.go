package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	"strconv"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
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
) {
	mux.Handle("/wbfeedback", ValidateOrCreateGigachatAccessKeyTmp(wbHandler(repository, processer)))
	mux.Handle("/api/v1/reviews", getReviewsHandler(repository))
}

func getReviewsHandler(repository Repository) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		cards, err := repository.GetDomainDataByFilter(filter)
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
		if err := json.NewEncoder(w).Encode(response); err != nil {
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

func ValidateOrCreateGigachatAccessKeyTmp(h http.Handler) http.Handler {
	now := time.Now()
	timeDifference := now.Unix() - gigachatAccessKeyCreationTime.Unix()
	if !(timeDifference <= 30*60) {
		url := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
		data := "scope=GIGACHAT_API_PERS"

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
		if err != nil {
			fmt.Printf("Error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("Accept", "application/json")

		rquid, err := uuid.NewRandom()
		if err != nil {
			fmt.Printf("Error generating UUID: %v", err)
		}
		req.Header.Set("RqUID", rquid.String())

		authKey := os.Getenv("GIGACHAT_AUTH_KEY")
		if authKey == "" {
			fmt.Printf("GIGACHAT_AUTH_KEY environment variable not set")
		}
		req.Header.Set("Authorization", "Basic "+authKey)

		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error sending request: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Response status is not okay: %v", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response: %v", err)
		}

		type Response struct {
			AccessToken string `json:"access_token"`
		}

		var response Response
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Printf("Error decoding response")
		}

		file, _ := os.Open(".env")
		defer file.Close()

		envVariables, _ := godotenv.Parse(file)
		envVariables["GIGACHAT_ACCESS_TOKEN"] = response.AccessToken

		godotenv.Write(envVariables, file.Name())
		godotenv.Load()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) })
}
