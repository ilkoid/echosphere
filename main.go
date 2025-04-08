package main

import (
	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	gigachat "echosphere/llm/gigachat"
	processer "echosphere/processer"
	"echosphere/sqlite"
	"fmt"
	"net/http"

	"github.com/joho/godotenv"
)

const (
	spreadsheetID   = "1TKuEZvJsnqfkolxXtQzDPPUxdtDEVrtgjvCrVhnAA2o"
	sheetRange      = "Лист1!A1:B"
	credentialsFile = "/etc/google_sheets_credentials.json"
	repoFilePath    = "repository/database.db"

	schemaSQL = `
	CREATE TABLE products (
		vendor_id INTEGER NOT NULL,
		wb_id INTEGER NOT NULL,
		name TEXT,
		description TEXT,
		PRIMARY KEY (vendor_id)
	);

	CREATE TABLE photos (
		id INTEGER PRIMARY KEY,
		byte_slice BLOB,
	);

	CREATE TABLE product_photos (
		product_vendor_code INTEGER NOT NULL,
		photo_id INTEGER NOT NULL,
		PRIMARY KEY (product_vendor_code, photo_id),
		FOREIGN KEY (product_vendor_code) REFERENCES products(vendor_id),
		FOREIGN KEY (photo_id) REFERENCES photos(id)
	);

	CREATE TABLE reviews (
		id INTEGER NOT NULL,
		published_at TIMESTAMP NOT NULL,
		rating INTEGER NOT NULL,
		text TEXT,
		published_response TEXT,
		suggested_response TEXT,
		mood TEXT,
		key_words TEXT,
		PRIMARY KEY (id)
	);

	CREATE TABLE review_of_product (
		review_id INTEGER NOT NULL,
		product_vendor_code INTEGER NOT NULL,
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
	sqlite.Init(schemaSQL)

	wb_reviews, _ := wb_api.GetFeedback(config)

	reviewProcesser := processer.NewLLMReviewProcesser(gigachat.GetGigachatAPI(), processer.LLMResponseGenerator{}, processer.LLMMoodRecognizer{}, processer.LLMKeywordFinder{})
	processedReviews, err := reviewProcesser.ProcessReviews(wb_reviews)
	if err != nil {
		fmt.Printf("Error was: %v\n", err)
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
