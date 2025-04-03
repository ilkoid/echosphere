package main

import (
	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	gigachat "echosphere/llm/gigachat"
	processer "echosphere/process"
	"fmt"
	"net/http"
)

const (
	spreadsheetID   = "1TKuEZvJsnqfkolxXtQzDPPUxdtDEVrtgjvCrVhnAA2o"
	sheetRange      = "Лист1!A1:B"
	credentialsFile = "/etc/google_sheets_credentials.json"
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

func main() {
	http.HandleFunc("/wbfeedback", handle_wb)

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
