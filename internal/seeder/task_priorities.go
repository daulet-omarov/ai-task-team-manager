package seeder

import (
	"log"
)

func (s *Seeder) SeedTaskPriorities() {
	_, err := s.db.Exec(`
        INSERT INTO task_priorities (id, name, code)
        VALUES 
            (1, 'Низкий', 'low'),
            (2, 'Средний', 'medium'),
            (3, 'Высокий',  'high'),
            (4, 'Неотложный', 'urgent')
        ON CONFLICT (id) DO NOTHING
    `)

	if err != nil {
		log.Printf("Failed to seed task_priorities: %v", err)
		return
	}

	log.Println("Task Priorities seeded!")
}
