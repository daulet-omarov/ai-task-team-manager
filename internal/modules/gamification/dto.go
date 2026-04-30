package gamification

import "time"

// ─── Reason constants (also used in tests) ───────────────────────────────────

const (
	ReasonTaskCompleted  = "task_completed"
	ReasonOnTimeBonus    = "on_time_bonus"
	ReasonQualityBonus   = "quality_bonus"
	ReasonStreakBonus     = "streak_bonus"
	ReasonStreakMilestone = "streak_milestone"
	ReasonComboBonus     = "combo_bonus"
	ReasonKudosReceived  = "kudos_received"
	ReasonReworkPenalty  = "rework_penalty"
)

// ─── Level thresholds ─────────────────────────────────────────────────────────

type LevelDef struct {
	Level int    `json:"level"`
	Name  string `json:"name"`
	Min   int    `json:"min"`
}

var Levels = []LevelDef{
	{1, "Novice", 0},
	{2, "Contributor", 100},
	{3, "Performer", 300},
	{4, "Expert", 700},
	{5, "Elite", 1500},
	{6, "Legend", 3000},
}

func ComputeLevel(pts int) int {
	level := 1
	for _, l := range Levels {
		if pts >= l.Min {
			level = l.Level
		}
	}
	return level
}

// ─── Request / response types ─────────────────────────────────────────────────

type GiveKudosRequest struct {
	ToUserID int64  `json:"to_user_id" validate:"required"`
	TaskID   *uint  `json:"task_id"`
	Message  string `json:"message"`
}

type LeaderboardEntry struct {
	Rank          int    `json:"rank"`
	UserID        int64  `json:"user_id"`
	FullName      string `json:"full_name"`
	Photo         string `json:"photo"`
	RollingPoints int    `json:"rolling_points"`
	TotalPoints   int    `json:"total_points"`
	CurrentLevel  int    `json:"current_level"`
	LevelName     string `json:"level_name"`
	CurrentStreak int    `json:"current_streak"`
}

type leaderboardRow struct {
	UserID        int64  `gorm:"column:user_id"`
	FullName      string `gorm:"column:full_name"`
	Photo         string `gorm:"column:photo"`
	RollingPoints int    `gorm:"column:rolling_points"`
	TotalPoints   int    `gorm:"column:total_points"`
	CurrentLevel  int    `gorm:"column:current_level"`
	CurrentStreak int    `gorm:"column:current_streak"`
}

type UserStatsResponse struct {
	UserID        int64  `json:"user_id"`
	TotalPoints   int    `json:"total_points"`
	Rolling30d    int    `json:"rolling_30d_points"`
	CurrentLevel  int    `json:"current_level"`
	LevelName     string `json:"level_name"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
	TodayPoints   int    `json:"today_points"`
	DailyCap      int    `json:"daily_cap"`
	DailyCapHit   bool   `json:"daily_cap_hit"`
	Levels        []LevelDef `json:"levels"`
}

type PointTransactionResponse struct {
	ID       uint      `json:"id"`
	TaskID   *uint     `json:"task_id"`
	Points   int       `json:"points"`
	Reason   string    `json:"reason"`
	EarnedAt time.Time `json:"earned_at"`
}

type KudosResponse struct {
	ID         uint      `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	ToUserID   int64     `json:"to_user_id"`
	TaskID     *uint     `json:"task_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type KudosStatusResponse struct {
	GivenThisWeek int `json:"given_this_week"`
	MaxPerWeek    int `json:"max_per_week"`
}

type receivedKudosRow struct {
	ID         uint      `gorm:"column:id"`
	FromUserID int64     `gorm:"column:from_user_id"`
	FromName   string    `gorm:"column:from_name"`
	FromPhoto  string    `gorm:"column:from_photo"`
	TaskID     *uint     `gorm:"column:task_id"`
	Message    string    `gorm:"column:message"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

type ReceivedKudosResponse struct {
	ID         uint      `json:"id"`
	FromUserID int64     `json:"from_user_id"`
	FromName   string    `json:"from_name"`
	FromPhoto  string    `json:"from_photo"`
	TaskID     *uint     `json:"task_id"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}
