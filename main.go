package main

import (
	wb "echosphere/apis/wb"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	// ozon "echosphere/apis/ozon"
)

func handle_wb(w http.ResponseWriter, r *http.Request) {
	wb_key := os.Getenv("WB_API_SAFE")

	isAnswered := "false"
	take := 2
	skip := 0

	url := fmt.Sprintf("https://feedbacks-api.wildberries.ru/api/v1/feedbacks?isAnswered=%s&take=%d&skip=%d", isAnswered, take, skip)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create request: %v", err), http.StatusInternalServerError)
		return
	}

	req.Header.Set("Authorization", "Bearer "+wb_key)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read response body: %v", err), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, fmt.Sprintf("API request failed with status: %s", resp.Status), http.StatusInternalServerError)
		return
	}

	var feedback wb.FeedbackResponse
	if err := json.Unmarshal(body, &feedback); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse JSON response: %v", err), http.StatusInternalServerError)
		return
	}

	// -------- CHECK HERE --------------
	condensed := wb.ConvertWbIntoCondensed(feedback)
	fmt.Printf("%+v\n", condensed)
	// -------- END CHECK ---------------

	w.WriteHeader(http.StatusOK)
}

func main() {
	http.HandleFunc("/wbfeedback", handle_wb)

	fmt.Println("Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
