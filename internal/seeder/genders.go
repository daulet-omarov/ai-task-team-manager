package seeder

import (
	"log"
)

func (s *Seeder) SeedGenders() {
	_, err := s.db.Exec(`
        INSERT INTO genders (id, name, code)
        VALUES 
            (1, 'Мужской', 'male'),
            (2, 'Женский', 'female'),
            (3, 'Другое',  'other')
        ON CONFLICT (id) DO NOTHING
    `)

	if err != nil {
		log.Printf("Failed to seed genders: %v", err)
		return
	}

	log.Println("Genders seeded!")
}
