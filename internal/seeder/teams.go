package seeder

import (
	"log"
)

func (s *Seeder) SeedTeams() {
	_, err := s.db.Exec(`
        INSERT INTO teams (id, name, code)
        VALUES 
            (1, 'Backend', 'backend'),
            (2, 'Frontend', 'frontend'),
            (3, 'Data Engineers',  'data_engineer'),
            (4, 'DevOps', 'devops'),
            (5, 'Designers', 'designer'),
            (6, 'Project Managers', 'pm')
        ON CONFLICT (id) DO NOTHING
    `)

	if err != nil {
		log.Printf("Failed to seed teams: %v", err)
		return
	}

	log.Println("Teams seeded!")
}
