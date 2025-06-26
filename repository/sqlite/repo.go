package sqlite

import (
	"fmt"
	"sync"

	"database/sql"
	"echosphere/repository/domain"
	server_structs "echosphere/server/structs"

	_ "modernc.org/sqlite"
)

type Repo struct {
	mu  sync.Mutex
	sql *sql.DB
}

func New(filePath string) *Repo {
	database, err := sql.Open("sqlite", filePath)
	if err != nil {
		fmt.Printf("Couldnt open database: %v", err)
		return nil
	}

	return &Repo{
		sql: database,
	}
}

func (r *Repo) Init(scheme string) error {
	if _, err := r.sql.Exec(scheme); err != nil {
		return err
	}
	return nil
}

func (r *Repo) IsProductExist(product domain.Product) (bool, error) {
	var exists bool
	if err := r.sql.QueryRow("SELECT EXISTS(SELECT 1 FROM products WHERE vendor_id = ?)", product.VendorId).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repo) IsReviewExist(review domain.Review) (bool, error) {
	var exists bool
	if err := r.sql.QueryRow("SELECT EXISTS(SELECT 1 FROM reviews WHERE id = ?)", review.Id).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *Repo) GetByVendorId(id int) (domain.Product, error) {
	var product domain.Product
	if err := r.sql.QueryRow("SELECT * FROM products WHERE vendor_id = ?", id).Scan(&product.VendorId, &product.WBId, &product.Name, &product.Description); err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

func (r *Repo) GetProductPhotos(product domain.Product) ([]domain.Photo, error) {
	var photos []domain.Photo

	rows, err := r.sql.Query("SELECT photo_id FROM product_photos WHERE product_vendor_code = ?", product.VendorId)
	if err != nil {
		return photos, err
	}
	defer rows.Close()

	var photoIds []int
	for rows.Next() {
		var photo_id int
		if err := rows.Scan(&photo_id); err != nil {
			return photos, err
		}
		photoIds = append(photoIds, photo_id)
	}
	if err = rows.Err(); err != nil {
		return photos, err
	}

	for _, photo_id := range photoIds {
		var photo domain.Photo
		if err := r.sql.QueryRow("SELECT byte_slice FROM photos WHERE id = ?", photo_id).Scan(&photo.ByteSlice); err != nil {
			return photos, err
		}
		photos = append(photos, photo)
	}

	return photos, nil
}

func (r *Repo) GetReview(reviewId int) (domain.Review, error) {
	var review domain.Review
	if err := r.sql.QueryRow("SELECT * FROM reviews WHERE id = ?", reviewId).Scan(&review.Id, &review.PublishedAt, &review.Rating, &review.Text, &review.PublishedResponse, &review.SuggestedResponse, &review.Mood, &review.KeyWords); err != nil {
		return domain.Review{}, err
	}

	return review, nil
}

func (r *Repo) GetProductOfReview(review domain.Review) (domain.Product, error) {
	var productId int
	if err := r.sql.QueryRow("SELECT product_vendor_code FROM review_of_product WHERE review_id = ?", review.Id).Scan(&productId); err != nil {
		return domain.Product{}, err
	}

	var product domain.Product
	if err := r.sql.QueryRow("SELECT * FROM products WHERE vendor_id = ?", productId).Scan(&product.VendorId, &product.WBId, &product.Name, &product.Description); err != nil {
		return domain.Product{}, err
	}

	return product, nil
}

func (r *Repo) AddProduct(product domain.Product, photos ...domain.Photo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.sql.Begin()
	if err != nil {
		return err
	}

	_, err = r.sql.Exec("INSERT INTO products (vendor_id, wb_id, name, description) VALUES (?, ?, ?, ?)", product.VendorId, product.WBId, product.Name, product.Description)
	if err != nil {
		return err
	}

	for _, photo := range photos {
		var photo_id int
		_, err = r.sql.Exec("INSERT INTO photos (byte_slice) VALUES (?)", photo.ByteSlice)
		if err != nil {
			tx.Rollback()
			return err
		}

		if err = r.sql.QueryRow("SELECT id FROM photos WHERE byte_slice = ?", photo.ByteSlice).Scan(&photo_id); err != nil {
			tx.Rollback()
			return err
		}

		_, err = r.sql.Exec("INSERT INTO product_photos (product_vendor_code, photo_id) VALUES (?, ?)", product.VendorId, photo_id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repo) AddReview(review domain.Review, product domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, err := r.sql.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = r.sql.Exec("INSERT INTO reviews (id, published_at, rating, text, published_response, suggested_response, mood, key_words) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", review.Id, review.PublishedAt, review.Rating, review.Text, review.PublishedResponse, review.SuggestedResponse, review.Mood, review.KeyWords)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = r.sql.Exec("INSERT INTO review_of_product (review_id, product_vendor_code) VALUES (?, ?)", review.Id, product.VendorId)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *Repo) GetCardsByFilter(filter server_structs.Filter) ([]server_structs.Card, error) {
	query := `
        SELECT r.id, r.published_at, r.rating, r.text, 
               r.published_response, r.suggested_response, r.mood, r.key_words,
               p.vendor_id, p.wb_id, p.name, p.description
        FROM reviews r
        JOIN review_of_product rop ON r.id = rop.review_id
        JOIN products p ON rop.product_vendor_code = p.vendor_id
        WHERE r.published_at >= datetime(?, 'unixepoch')
    `
	args := []interface{}{filter.DateFrom}

	if filter.Rating != nil {
		query += " AND r.rating >= ?"
		args = append(args, *filter.Rating)
	}

	if filter.VendorId != "" {
		query += " AND p.vendor_id = ?"
		args = append(args, filter.VendorId)
	}

	rows, err := r.sql.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer rows.Close()

	var cards []server_structs.Card

	for rows.Next() {
		var review domain.Review
		var product domain.Product

		err := rows.Scan(
			&review.Id,
			&review.PublishedAt,
			&review.Rating,
			&review.Text,
			&review.PublishedResponse,
			&review.SuggestedResponse,
			&review.Mood,
			&review.KeyWords,
			&product.VendorId,
			&product.WBId,
			&product.Name,
			&product.Description,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card row: %w", err)
		}

		photos, err := r.GetProductPhotos(product)
		if err != nil {
			return nil, fmt.Errorf("failed to get product photos: %w", err)
		}

		cards = append(cards, server_structs.Card{
			Review:  review,
			Product: product,
			Photos:  photos,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return cards, nil
}

func (r *Repo) UpdateReviews(updates []server_structs.MarketplaceResponse) error {
	tx, err := r.sql.Begin()
	if err != nil {
		return err
	}

	for _, update := range updates {
		_, err := r.sql.Exec("UPDATE reviews SET published_response = ? WHERE id = ?", update.Text, update.Id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
