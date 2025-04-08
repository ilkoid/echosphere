package repository

import (
	"echosphere/repository/domain"
)

type Repo interface {
	GetByVendorId(id int) (domain.Product, error)
	GetProductPhotos(product domain.Product) ([]domain.Photo, error)
	GetReview(reviewId int) (domain.Review, error)
	GetProductOfReview(review domain.Review) (domain.Product, error)

	AddProduct(product domain.Product, photos ...domain.Photo) error
	AddReview(review domain.Review) error
}
