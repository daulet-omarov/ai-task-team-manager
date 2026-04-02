package seeder

import (
	"log"
)

func (s *Seeder) SeedStatuses() {
	_, err := s.db.Exec(`
        INSERT INTO statuses (id, name, code)
        VALUES 
            (1, 'BACKLOG', 'backlog'),
            (2, 'BLOCKED', 'blocked'),
            (3, 'NEED REWORK', 'need_rework'),
            (4, 'IN PROGRESS', 'in_progress'),
            (5, 'CODE REVIEW', 'code_review'),
            (6, 'TEST', 'test'),
            (7, 'READY FOR DEPLOY', 'ready_for_deploy'),
            (8, 'IN PRODUCTION', 'in_production'),
            (9, 'DONE', 'done')
        ON CONFLICT (id) DO NOTHING
    `)

	if err != nil {
		log.Printf("Failed to seed statuses: %v", err)
		return
	}

	log.Println("Statuses seeded!")
}
