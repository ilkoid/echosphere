package main

import (
	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	gigachat "echosphere/llm/gigachat"
	processer "echosphere/processer"
	"echosphere/repository/domain"
	"echosphere/sqlite"
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
)

const (
	spreadsheetID   = "1TKuEZvJsnqfkolxXtQzDPPUxdtDEVrtgjvCrVhnAA2o"
	sheetRange      = "Лист1!A1:B"
	credentialsFile = "/etc/google_sheets_credentials.json"
	repoFilePath    = "sqlite/database.db"

	schemaSQL = `
	CREATE TABLE IF NOT EXISTS products(
		vendor_id TEXT PRIMARY KEY,
		wb_id INTEGER NOT NULL,
		name TEXT,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS photos (
		id INTEGER PRIMARY KEY,
		byte_slice BLOB
	);

	CREATE TABLE IF NOT EXISTS product_photos (
		product_vendor_code TEXT NOT NULL,
		photo_id INTEGER NOT NULL,
		PRIMARY KEY (product_vendor_code, photo_id),
		FOREIGN KEY (product_vendor_code) REFERENCES products(vendor_id),
		FOREIGN KEY (photo_id) REFERENCES photos(id)
	);

	CREATE TABLE IF NOT EXISTS reviews (
		id TEXT PRIMARY KEY,
		published_at TIMESTAMP NOT NULL,
		rating INTEGER NOT NULL,
		text TEXT,
		published_response TEXT,
		suggested_response TEXT,
		mood TEXT,
		key_words TEXT
	);

	CREATE TABLE IF NOT EXISTS review_of_product (
		review_id TEXT NOT NULL,
		product_vendor_code TEXT NOT NULL,
		PRIMARY KEY (review_id, product_vendor_code),
		FOREIGN KEY (review_id) REFERENCES reviews(id),
		FOREIGN KEY (product_vendor_code) REFERENCES products(vendor_id)
	);
	`
)

func handle_wb(w http.ResponseWriter, r *http.Request) {
	wb_api := wb.WBAPI{}
	config := apis.FeedbackRequestConfig{
		IsAnswered: false,
		Take:       10,
		Skip:       0,
		DateFrom:   nil,
		DateTo:     nil,
	}

	sqlite := sqlite.New(repoFilePath)
	err := sqlite.Init(schemaSQL)
	if err != nil {
		fmt.Printf("Couldnt init the db: %v\n", err)
	}

	wb_reviews, _ := wb_api.GetFeedback(config)

	reviewProcesser := processer.NewLLMReviewProcesser(gigachat.GetGigachatAPI(), processer.LLMResponseGenerator{}, processer.LLMMoodRecognizer{}, processer.LLMKeywordFinder{})
	processedReviews, err := reviewProcesser.ProcessReviews(wb_reviews)
	if err != nil {
		fmt.Printf("Error was: %v\n", err)
	}

	for _, review := range processedReviews {
		err := sqlite.AddProduct(
			domain.Product{
				VendorId:    review.Product.VendorCode,
				WBId:        review.Product.ID,
				Name:        review.Product.Name,
				Description: review.Product.Description,
			},
		)
		if err != nil {
			fmt.Printf("Couldnt add the product to the db: %v\n", err)
		}

		err = sqlite.AddReview(
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
}

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {
	http.HandleFunc("/wbfeedback", handle_wb)

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
