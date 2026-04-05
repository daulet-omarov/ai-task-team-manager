package seeder

import (
	"log"
)

func (s *Seeder) SeedTaskDifficulties() {
	_, err := s.db.Exec(`
        INSERT INTO difficulties (id, name, code)
        VALUES 
            (1, 'Легкий', 'easy'),
            (2, 'Средний', 'medium'),
            (3, 'Сложный',  'hard')
        ON CONFLICT (id) DO NOTHING
    `)

	if err != nil {
		log.Printf("Failed to seed task_difficulties: %v", err)
		return
	}

	log.Println("Task Difficulties seeded!")
}
