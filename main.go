package main

import (
	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	gigachat "echosphere/llm/gigachat"
	"fmt"
	"net/http"
)

const (
	spreadsheetID   = "1TKuEZvJsnqfkolxXtQzDPPUxdtDEVrtgjvCrVhnAA2o"
	sheetRange      = "Лист1!A1:B"
	credentialsFile = "/etc/google_sheets_credentials.json"
)

func handle_wb(w http.ResponseWriter, r *http.Request) {
	// http.Error(w, fmt.Sprintf("Failed to parse JSON response: %v", err), http.StatusInternalServerError)

	wb_api := wb.WBAPI{}
	config := apis.FeedbackRequestConfig{
		IsAnswered: false,
		Take:       10,
		Skip:       0,
		DateFrom:   nil,
		DateTo:     nil,
	}

	wb_reviews, _ := wb_api.GetFeedback(config)

	largeLanguageModel := gigachat.GigachatAPI{}

	// fmt.Printf("Before: %v\n", wb_reviews)

	wb_reviews, err := largeLanguageModel.MakeResponse(wb_reviews)
	if err != nil {
		fmt.Printf("Error was: %v\n", err)
	}

	// fmt.Printf("After: %v\n", wb_reviews)

	values := google_sheets.PrepareDataForSheets(wb_reviews)
	if err := google_sheets.WriteReviewsToSheet(values, spreadsheetID, sheetRange, credentialsFile); err != nil {
		http.Error(w, fmt.Sprintf("Failed to write to Google Sheets: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/wbfeedback", handle_wb)

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
