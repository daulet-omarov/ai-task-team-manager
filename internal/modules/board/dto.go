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
