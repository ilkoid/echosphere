package main

import (
	gigachat "echosphere/llm/gigachat"
	processer "echosphere/processer"
	"echosphere/repository/sqlite"
	"echosphere/server"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const (
	repoFilePath    = "repository/sqlite/database.db"

	schemaSQL = `
	CREATE TABLE IF NOT EXISTS products(
		vendor_id TEXT PRIMARY KEY,
		wb_id INTEGER NOT NULL,
		name TEXT,
		description TEXT
	);

	CREATE TABLE IF NOT EXISTS photos (
		id INTEGER PRIMARY KEY,
		byte_slice BLOB
	);

	CREATE TABLE IF NOT EXISTS product_photos (
		product_vendor_code TEXT NOT NULL,
		photo_id INTEGER NOT NULL,
		PRIMARY KEY (product_vendor_code, photo_id),
		FOREIGN KEY (product_vendor_code) REFERENCES products(vendor_id) ON DELETE CASCADE,
		FOREIGN KEY (photo_id) REFERENCES photos(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS reviews (
		id TEXT PRIMARY KEY,
		published_at TIMESTAMP NOT NULL,
		rating INTEGER NOT NULL,
		text TEXT,
		published_response TEXT,
		suggested_response TEXT,
		mood TEXT,
		key_words TEXT
	);

	CREATE TABLE IF NOT EXISTS review_of_product (
		review_id TEXT NOT NULL,
		product_vendor_code TEXT NOT NULL,
		PRIMARY KEY (review_id, product_vendor_code),
		FOREIGN KEY (review_id) REFERENCES reviews(id) ON DELETE CASCADE,
		FOREIGN KEY (product_vendor_code) REFERENCES products(vendor_id) ON DELETE CASCADE
	);
	`
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found")
	}
}

func main() {
	// http.HandleFunc("/wbfeedback", handle_wb)
	//
	sqlite := sqlite.New(repoFilePath)
	if sqlite == nil {
		fmt.Printf("Couldnt open the db\n")
	}
	err := sqlite.Init(schemaSQL)
	if err != nil {
		fmt.Printf("Couldnt init the db: %v\n", err)
	}

	reviewProcesser := processer.NewLLMReviewProcesser(
		gigachat.GetGigachatAPI(),
		processer.LLMResponseGenerator{},
		processer.LLMMoodRecognizer{},
		processer.LLMKeywordFinder{},
	)

	srv := server.New(reviewProcesser, *sqlite)

	httpSrv := http.Server{
		Addr:    net.JoinHostPort("localhost", "8080"),
		Handler: srv,
	}

	fmt.Printf("listening on %s\n", httpSrv.Addr)
	if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
	}
}
