package task

type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=200"`
	Description string `json:"description"`
	PriorityID  uint   `json:"priority_id"`
	AssigneeID  uint   `json:"assignee_id"` // employee ID
}

type UpdateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	StatusID    uint   `json:"status_id"`
	PriorityID  uint   `json:"priority_id"`
	AssigneeID  uint   `json:"assignee_id"` // employee ID
	TimeSpent   uint   `json:"time_spent"`
}

type TaskResponse struct {
	ID          uint   `json:"id"`
	BoardID     uint   `json:"board_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	StatusID    uint   `json:"status_id"`
	PriorityID  uint   `json:"priority_id"`
	AssigneeID  uint   `json:"assignee_id"`
	ReporterID  uint   `json:"reporter_id"`
	TimeSpent   uint   `json:"time_spent"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
