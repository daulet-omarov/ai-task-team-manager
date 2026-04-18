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
	UserID   int64  `json:"user_id"`
	Role     string `json:"role"`
	FullName string `json:"full_name"`
	Photo    string `json:"photo"`
	Email    string `json:"email"`
}

type CreateStatusRequest struct {
	Title string `json:"title" validate:"required,min=1,max=100"`
}

type StatusResponse struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Code     string `json:"code"`
	Position int    `json:"position"`
}

type ReorderStatusesRequest struct {
	Statuses []StatusPosition `json:"statuses" validate:"required,min=1"`
}

type StatusPosition struct {
	StatusID uint `json:"status_id" validate:"required"`
	Position int  `json:"position" validate:"required"`
}
