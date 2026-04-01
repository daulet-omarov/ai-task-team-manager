package seeder

import (
	"database/sql"
	"log"
)

type Seeder struct {
	db *sql.DB
}

func New(db *sql.DB) *Seeder {
	return &Seeder{db: db}
}

func (s *Seeder) Run() {
	s.SeedGenders()

	log.Println("Seeding completed!")
}
