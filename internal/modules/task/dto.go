package task

import (
	"mime/multipart"
	"time"
)

type CreateTaskRequest struct {
	Title        string
	Description  string
	PriorityID   uint
	DifficultyID *uint
	AssigneeID   uint
	TesterID     uint
	Files        []*multipart.FileHeader // optional attachments
}

type UpdateTaskRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	StatusID     uint   `json:"status_id"`
	PriorityID   uint   `json:"priority_id"`
	DifficultyID *uint  `json:"difficulty_id"`
	AssigneeID   uint   `json:"assignee_id"` // employee ID
	TesterID     uint   `json:"tester_id"`   // employee ID
	TimeSpent    uint   `json:"time_spent"`
}

type AttachmentInfo struct {
	ID       uint   `json:"id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	URL      string `json:"url"`
}

type CommentInfo struct {
	ID        uint      `json:"id"`
	AuthorID  uint      `json:"author_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskResponse struct {
	ID           uint             `json:"id"`
	BoardID      uint             `json:"board_id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	StatusID     uint             `json:"status_id"`
	PriorityID   uint             `json:"priority_id"`
	DifficultyID *uint            `json:"difficulty_id"`
	AssigneeID   uint             `json:"assignee_id"`
	TesterID     uint             `json:"tester_id"`
	ReporterID   uint             `json:"reporter_id"`
	TimeSpent    uint             `json:"time_spent"`
	Attachments  []AttachmentInfo `json:"attachments"`
	Comments     []CommentInfo    `json:"comments"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
}
