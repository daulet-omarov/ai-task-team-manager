package invite

type CreateInviteRequest struct {
	UserID int64 `json:"user_id" validate:"required"`
}

type InviteResponse struct {
	ID        uint   `json:"id"`
	BoardID   uint   `json:"board_id"`
	BoardName string `json:"board_name"`
	InviterID int64  `json:"inviter_id"`
	InviteeID int64  `json:"invitee_id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
