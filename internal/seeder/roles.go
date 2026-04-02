package seeder

import (
	"log"
)

func (s *Seeder) SeedRoles() {
	_, err := s.db.Exec(`
        INSERT INTO roles (id, name, code)
        VALUES 
            (1, 'Супер админ', 'super_admin'),
            (2, 'Админ', 'admin'),
            (3, 'Пользователь',  'user')
        ON CONFLICT (id) DO NOTHING
    `)

	if err != nil {
		log.Printf("Failed to seed roles: %v", err)
		return
	}

	log.Println("Roles seeded!")
}
