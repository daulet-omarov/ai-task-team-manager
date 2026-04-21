package comment

import "time"

type CommentWithAuthor struct {
	ID             uint
	TaskID         uint
	AuthorID       uint
	AuthorFullName string
	AuthorPhoto    string
	Content        string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=1"`
}

type CommentResponse struct {
	ID             uint      `json:"id"`
	TaskID         uint      `json:"task_id"`
	AuthorID       uint      `json:"author_id"`
	AuthorFullName string    `json:"author_full_name"`
	AuthorPhoto    string    `json:"author_photo"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
