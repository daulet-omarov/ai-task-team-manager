package models

import "time"

type BoardChatMessage struct {
	ID          uint `gorm:"primaryKey"`
	BoardID     uint `gorm:"not null"`
	AuthorID    uint `gorm:"not null"`
	ReplyToID   *uint
	Text        string `gorm:"type:text"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Author      Employee
	ReplyTo     *BoardChatMessage
	Attachments []BoardChatAttachment `gorm:"foreignKey:MessageID"`
	Poll        *BoardPoll            `gorm:"foreignKey:MessageID"`
}

type BoardChatAttachment struct {
	ID        uint   `gorm:"primaryKey"`
	MessageID uint   `gorm:"not null"`
	FilePath  string `gorm:"not null"`
	FileName  string `gorm:"not null"`
	FileSize  int    `gorm:"not null"`
	MimeType  string
	CreatedAt time.Time
}

type BoardPoll struct {
	ID        uint   `gorm:"primaryKey"`
	MessageID uint   `gorm:"not null"`
	Question  string `gorm:"not null"`
	CreatedAt time.Time
	Options   []BoardPollOption `gorm:"foreignKey:PollID"`
}

type BoardPollOption struct {
	ID     uint            `gorm:"primaryKey"`
	PollID uint            `gorm:"not null"`
	Text   string          `gorm:"not null"`
	Votes  []BoardPollVote `gorm:"foreignKey:OptionID"`
}

type BoardPollVote struct {
	ID         uint `gorm:"primaryKey"`
	OptionID   uint `gorm:"not null"`
	EmployeeID uint `gorm:"not null"`
	CreatedAt  time.Time
}
