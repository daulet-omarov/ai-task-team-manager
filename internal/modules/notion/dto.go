package notion

type ImportRequest struct {
	Token      string `json:"token"`
	DatabaseID string `json:"database_id"`
	// BoardID is optional; when 0 a new board is created from the Notion database title.
	BoardID uint `json:"board_id"`
}

type ImportResult struct {
	BoardID      uint     `json:"board_id"`
	BoardName    string   `json:"board_name"`
	TasksCreated int      `json:"tasks_created"`
	Skipped      int      `json:"skipped"`
	Errors       []string `json:"errors,omitempty"`
}
