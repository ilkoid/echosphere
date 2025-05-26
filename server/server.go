package server

import (
	"echosphere/apis"
	wb "echosphere/apis/wb"
	"echosphere/google_sheets"
	"echosphere/processer"
	"echosphere/repository/domain"

	"bytes"
	"context"
	"crypto/tls"
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

type ReviewProcesser interface {
	ProcessReviews(reviews []apis.ReviewCondensed) ([]processer.ReviewProcessed, error)
}

type Repository interface {
	IsProductExist(product domain.Product) (bool, error)
	IsReviewExist(review domain.Review) (bool, error)

	GetByVendorId(id int) (domain.Product, error)
	GetProductPhotos(product domain.Product) ([]domain.Photo, error)
	GetReview(reviewId int) (domain.Review, error)
	GetProductOfReview(review domain.Review) (domain.Product, error)
	GetCardsByFilter(filter Filter) ([]Card, error)

	AddProduct(product domain.Product, photos ...domain.Photo) error
	AddReview(review domain.Review, product domain.Product) error

	UpdateReviews(updates []ReviewUpdate) error
}

type Card struct {
	Review  domain.Review
	Product domain.Product
	Photos  []domain.Photo
}

type Filter struct {
	DateFrom int64
	Rating   *int
	VendorId string
}

type ReviewUpdate struct {
	Id                string `json:"id"`
	PublishedResponse string `json:"published_response"`
}

type MarketplaceAPI interface {
	GetFeedback(config apis.FeedbackRequestConfig) ([]apis.ReviewCondensed, error)
}

func New(
	processer ReviewProcesser,
	repository Repository,
) http.Handler {
	mux := http.NewServeMux()

	addRoutes(
		mux,
		processer,
		repository,
	)

	var handler http.Handler = mux
	handler = AuthMiddleware(handler)
	return handler
}

func AuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("User")
		password := r.Header.Get("Password")

		if user == "admin" && password == "secret" {
			handler.ServeHTTP(w, r)
		} else {
			http.Error(w, "Error 401: Unauthorised", http.StatusUnauthorized)
		}
	})
}

func RunProcesses(context context.Context, processer ReviewProcesser, repository Repository) {
	type Config struct {
		IsAnswered             bool `yaml:"is_answered"`
		Take                   int  `yaml:"take"`
		Skip                   int  `yaml:"skip"`
		UpdateDatabaseInterval int  `yaml:"update_database_interval"`
	}

	file, _ := os.Open("config.yaml")

	var conf Config
	bytes, _ := io.ReadAll(file)
	yaml.Unmarshal(bytes, &conf)

	wbApi := wb.WBAPI{}

	// first initilial db update
	updateDatabase(&wbApi, processer, repository)

	select {
	case <-context.Done():
		file.Close()
		return
	case <-time.After(time.Duration(conf.UpdateDatabaseInterval) * time.Minute):
		updateDatabase(&wbApi, processer, repository)
	}
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func updateDatabase(api MarketplaceAPI, processer ReviewProcesser, repository Repository) {
	type Config struct {
		IsAnswered             bool `yaml:"is_answered"`
		Take                   int  `yaml:"take"`
		Skip                   int  `yaml:"skip"`
		UpdateDatabaseInterval int  `yaml:"update_database_interval"`
	}

	file, _ := os.Open("config.yaml")

	var conf Config
	bytes, _ := io.ReadAll(file)
	yaml.Unmarshal(bytes, &conf)

	validateOrCreateGigachatAccessKey()

	now := time.Now()
	dateFrom := now.Add(-(time.Duration(conf.UpdateDatabaseInterval) * time.Minute))
	dateTo := now

	config := apis.FeedbackRequestConfig{
		IsAnswered: conf.IsAnswered,
		Take:       conf.Take,
		Skip:       conf.Skip,
		DateFrom:   &dateFrom,
		DateTo:     &dateTo,
	}

	reviews, err := api.GetFeedback(config)
	if err != nil {
		fmt.Println(err)
	}
	processedReviews, err := processer.ProcessReviews(reviews)
	if err != nil {
		fmt.Println(err)
	}

	err = writeToDatabase(processedReviews, repository)
	if err != nil {
		fmt.Println(err)
	}

	// soon to be deleted...
	values := google_sheets.PrepareDataForSheets(processedReviews)
	if err := google_sheets.WriteReviewsToSheet(values, spreadsheetID, sheetRange, credentialsFile); err != nil {
		fmt.Println(err)
	}
}

func validateOrCreateGigachatAccessKey() {
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
}

func writeToDatabase(processedReviews []processer.ReviewProcessed, repository Repository) error {
	for _, review := range processedReviews {
		product, photos := ConvertIntoDomainProduct(review)
		domainReview := ConvertIntoDomainReview(review)

		exists, err := repository.IsProductExist(product)
		if err != nil {
			return err
		}
		if !exists {
			err = repository.AddProduct(product, photos...)
			if err != nil {
				return err
			}
		}

		exists, err = repository.IsReviewExist(domainReview)
		if err != nil {
			return err
		}
		if !exists {
			err = repository.AddReview(domainReview, product)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
