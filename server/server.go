package server

import (
	"echosphere/apis"
	"echosphere/processer"
	"echosphere/repository/domain"
	"encoding/json"
	"fmt"
	"net/http"
)

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

	AddProduct(product domain.Product, photos ...domain.Photo) error
	AddReview(review domain.Review, product domain.Product) error
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
	return handler
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
