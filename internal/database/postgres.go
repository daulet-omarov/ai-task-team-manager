package database

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func NewPostgres() *sql.DB {
	dsn := "postgres://postgres:postgres@10.10.72.55:5432/myapp?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	return db
}
