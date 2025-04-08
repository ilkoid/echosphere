package sqlite

import (
	"database/sql"
	"echosphere/repository/domain"
)

const (
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

type Repo struct {
	sql *sql.DB
}

func New(filePath string) *Repo {
	database, err := sql.Open("sqlite", filePath)
	if err != nil {
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
		if err = r.sql.QueryRow("INSERT INTO photos (byte_slice) VALUES (?)", photo.ByteSlice).Scan(&photo_id); err != nil {
			return err
		}
		_, err = r.sql.Exec("INSERT INTO products_photos (product_vendor_code, photo_id) VALUES (?, ?)", product.VendorId, photo_id)
	}

	return tx.Commit()
}

func (r *Repo) AddReview(review domain.Review, product domain.Product) error {
	tx, err := r.sql.Begin()
	if err != nil {
		return err
	}

	_, err = r.sql.Exec("INSERT INTO reviews (id, published_at, rating, text, published_response, suggested_response, mood, key_words) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", review.Id, review.PublishedAt, review.Rating, review.Text, review.PublishedResponse, review.SuggestedResponse, review.Mood, review.KeyWords)
	if err != nil {
		return err
	}

	_, err = r.sql.Exec("INSERT INTO review_of_product (review_id, product_vendor_code) VALUES (?, ?)", review.Id, product.VendorId)

	return tx.Commit()
}
