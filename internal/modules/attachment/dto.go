package attachment

import "time"

type AttachmentResponse struct {
	ID        uint      `json:"id"`
	TaskID    uint      `json:"task_id"`
	FileName  string    `json:"file_name"`
	FileSize  int       `json:"file_size"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
}
