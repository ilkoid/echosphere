package server

import (
	"bytes"
	"crypto/tls"
	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

var gigachatAccessKeyCreationTime time.Time = time.Unix(0, 0)

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
	mux.Handle("/wbfeedback", ValidateOrCreateGigachatAccessKey(wbHandler(repository, processer)))
}

func wbHandler(repository Repository, processer ReviewProcesser) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wbApi := wb.WBAPI{}
		file, _ := os.Open("config.yaml")
		bytes, _ := io.ReadAll(file)
		type Config struct {
			IsAnswered bool `yaml:"is_answered"`
			Take       int  `yaml:"take"`
			Skip       int  `yaml:"skip"`
		}
		var conf Config
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

		// for _, review := range processedReviews {
		// 	product, photos := ConvertIntoDomainProduct(review)
		// 	domainReview := ConvertIntoDomainReview(review)
		// 	err := repository.AddProduct(
		// 		product,
		// 		photos...,
		// 	)
		// 	if err != nil {
		// 		fmt.Printf("Couldnt add the product to the db: %v\n", err)
		// 	}

		// 	err = repository.AddReview(
		// 		domainReview,
		// 		product,
		// 	)
		// 	if err != nil {
		// 		fmt.Printf("Couldnt add the review to the db: %v\n", err)
		// 	}
		// }

		values := google_sheets.PrepareDataForSheets(processedReviews)
		if err := google_sheets.WriteReviewsToSheet(values, spreadsheetID, sheetRange, credentialsFile); err != nil {
			http.Error(w, fmt.Sprintf("Failed to write to Google Sheets: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func ValidateOrCreateGigachatAccessKey(h http.Handler) http.Handler {
	now := time.Now()
	timeDifference := now.Unix() - gigachatAccessKeyCreationTime.Unix()
	if !(timeDifference <= 30*60) {
		url := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
		data := "scope=GIGACHAT_API_PERS"

		req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
		if err != nil {
			fmt.Errorf("Error creating request: %v", err)
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
			fmt.Printf("Please forgive: %v", resp.StatusCode)
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
		envVariables, _ := godotenv.Parse(file)
		envVariables["GIGACHAT_ACCESS_TOKEN"] = response.AccessToken
		godotenv.Write(envVariables, file.Name())
		godotenv.Load()
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { h.ServeHTTP(w, r) })
}
