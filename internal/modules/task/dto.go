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
	DueDate      string `json:"due_date"` // RFC3339; empty means no change
}

type EmployeeInfo struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Photo    string `json:"photo"`
}

type AttachmentInfo struct {
	ID       uint   `json:"id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	URL      string `json:"url"`
}

type CommentInfo struct {
	ID             uint      `json:"id"`
	AuthorID       uint      `json:"author_id"`
	AuthorFullName string    `json:"author_full_name"`
	AuthorPhoto    string    `json:"author_photo"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
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
	Assignee     *EmployeeInfo    `json:"assignee"`
	TesterID     uint             `json:"tester_id"`
	Tester       *EmployeeInfo    `json:"tester"`
	ReporterID   uint             `json:"reporter_id"`
	Reporter     *EmployeeInfo    `json:"reporter"`
	TimeSpent    uint             `json:"time_spent"`
	DueDate      *string          `json:"due_date,omitempty"`
	Attachments  []AttachmentInfo `json:"attachments"`
	Comments     []CommentInfo    `json:"comments"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
}
