package auth

import "database/sql"

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(email, password string) error {
	query := `
	INSERT INTO users (email, password)
	VALUES ($1, $2)
	`

	_, err := r.db.Exec(query, email, password)
	return err
}

func (r *Repository) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, email, password
	FROM users
	WHERE email=$1
	`

	user := &User{}

	err := r.db.QueryRow(query, email).
		Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *Repository) DeleteUser(id int64) error {
	query := `DELETE FROM users WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}
