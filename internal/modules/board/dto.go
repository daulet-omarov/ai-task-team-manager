package board

type CreateBoardRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
}

type BoardResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	MemberCount int    `json:"memberCount"`
	IsOwner     bool   `json:"isOwner"`
	OwnerID     int64  `json:"ownerId"`
}

type DashboardResponse struct {
	Boards       []BoardResponse `json:"boards"`
	IsFirstLogin bool            `json:"isFirstLogin"`
}

type MemberResponse struct {
	BoardMemberID uint   `json:"board_member_id"`
	UserID        int64  `json:"user_id"`
	Role          string `json:"role"`
	FullName      string `json:"full_name"`
	Photo         string `json:"photo"`
	Email         string `json:"email"`
}

type CreateStatusRequest struct {
	Title   string `json:"title"    validate:"required,min=1,max=100"`
	BoardID uint   `json:"board_id" validate:"required"`
	Colour  string `json:"colour"`
}

type UpdateStatusRequest struct {
	Title  string `json:"title"`
	Colour string `json:"colour"`
}

type StatusResponse struct {
	BoardStatusID uint   `json:"board_status_id"` // id in board_statuses — use for reorder/delete
	StatusID      uint   `json:"status_id"`
	Name          string `json:"name"`
	Code          string `json:"code"`
	Position      int    `json:"position"`
	Colour        string `json:"colour"`
}

type ReorderStatusesRequest struct {
	Statuses []StatusPosition `json:"statuses" validate:"required,min=1"`
}

type StatusPosition struct {
	BoardStatusID uint `json:"board_status_id" validate:"required"`
	Position      int  `json:"position"        validate:"required"`
}
