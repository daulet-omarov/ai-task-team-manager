package chat

import "time"

// ── Requests ──────────────────────────────────────────────────────────────────

type SendMessageRequest struct {
	Text      string `form:"text"`
	ReplyToID *uint  `form:"reply_to_id"`
}

type CreatePollRequest struct {
	Question string   `json:"question" validate:"required"`
	Options  []string `json:"options"  validate:"required,min=2,dive,required"`
}

type VoteRequest struct {
	OptionID uint `json:"option_id" validate:"required"`
}

// ── Responses ─────────────────────────────────────────────────────────────────

type AuthorInfo struct {
	ID       uint   `json:"id"`
	FullName string `json:"full_name"`
	Photo    string `json:"photo"`
}

type AttachmentResponse struct {
	ID       uint   `json:"id"`
	FileName string `json:"file_name"`
	FileSize int    `json:"file_size"`
	MimeType string `json:"mime_type"`
	URL      string `json:"url"`
}

type PollOptionResponse struct {
	ID        uint         `json:"id"`
	Text      string       `json:"text"`
	VoteCount int          `json:"vote_count"`
	Voters    []AuthorInfo `json:"voters"`
}

type PollResponse struct {
	ID       uint                 `json:"id"`
	Question string               `json:"question"`
	Options  []PollOptionResponse `json:"options"`
}

type ReplyToResponse struct {
	ID     uint        `json:"id"`
	Author *AuthorInfo `json:"author"`
	Text   string      `json:"text"`
}

type MessageResponse struct {
	ID          uint                 `json:"id"`
	BoardID     uint                 `json:"board_id"`
	Author      AuthorInfo           `json:"author"`
	Text        string               `json:"text"`
	ReplyTo     *ReplyToResponse     `json:"reply_to,omitempty"`
	Attachments []AttachmentResponse `json:"attachments"`
	Poll        *PollResponse        `json:"poll,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}
