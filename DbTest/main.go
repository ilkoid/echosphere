package main

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite" 
)

func main() {
	db, err := sql.Open("sqlite", "test.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS feedback (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Status TEXT NOT NULL,
			Rating INTEGER,
			Text TEXT,
			Response TEXT
		);`)
	if err != nil {
		log.Fatalf("Ошибка при создании таблицы: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO feedback (Status, Rating, Text, Response) VALUES
			('none', 5, 'Отличный сервис, Happy Happy Happy!', 'Спасибо за отзыв! Наш Слон'),
			('none', 4, 'Быстрая доставка, но кот упал с вентилятора', 'Спасибо за отзыв!'),
			('none', 5, 'Никитос, Извинись', 'Приносим извинения')
	`)
	if err != nil {
		log.Fatalf("Ошибка при добавлении данных: %v", err)
	}

	log.Println("Данные успешно добавлены в таблицу feedback")
}