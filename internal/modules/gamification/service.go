package gamification

import (
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"gorm.io/gorm"
)

const (
	DailyCap       = 100
	MaxKudosPerWeek = 3
)

// difficultyComplexity maps difficulty codes to a 0-10 complexity score used in
// the multiplier formula: multiplier = 1 + complexity/10.
var difficultyComplexity = map[string]float64{
	"easy":   3,
	"medium": 5,
	"hard":   10,
}

// TODO: Role-based multipliers (Tester bonus for bug-catch tasks, Reporter bonus
// for quickly-accepted submissions) are not implemented because the current schema
// uses auth roles (super_admin/admin/user) rather than task-domain roles. Extend
// the employee schema with a domain role field to enable this.

type Service struct {
	repo *Repository
	db   *gorm.DB
}

func NewService(repo *Repository, db *gorm.DB) *Service {
	return &Service{repo: repo, db: db}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

type budget struct{ remaining int }

func newBudget(cap, used int) *budget {
	r := cap - used
	if r < 0 {
		r = 0
	}
	return &budget{remaining: r}
}

func (b *budget) take(pts int) int {
	if b.remaining <= 0 || pts <= 0 {
		return 0
	}
	if pts > b.remaining {
		pts = b.remaining
	}
	b.remaining -= pts
	return pts
}

func (b *budget) exhausted() bool { return b.remaining <= 0 }

func jsonMeta(m map[string]interface{}) json.RawMessage {
	bytes, _ := json.Marshal(m)
	return bytes
}

func complexityMultiplier(code string) float64 {
	c, ok := difficultyComplexity[code]
	if !ok {
		return 1.0
	}
	return 1.0 + c/10.0
}

// ─── OnTaskCompleted ──────────────────────────────────────────────────────────

// OnTaskCompleted awards points when a task transitions to is_completed status.
// developerID is the employee (= user) ID of the task's assignee.
func (s *Service) OnTaskCompleted(developerUserID int64, task *models.Task) error {
	if developerUserID == 0 {
		return nil // unassigned task — no one to award
	}

	now := time.Now().UTC()

	// 1. Enforce 5-minute minimum: ignore tasks closed too quickly.
	if now.Sub(task.CreatedAt) < 5*time.Minute {
		return nil
	}

	// 2. Query today's already-earned points to respect the daily cap.
	todayPts, err := s.repo.SumPointsToday(developerUserID)
	if err != nil {
		return err
	}
	if todayPts >= DailyCap {
		return nil // cap already hit
	}
	bgt := newBudget(DailyCap, todayPts)

	// 3. Resolve complexity multiplier from difficulty.
	mult := 1.0
	if task.DifficultyID != nil {
		code, err := s.repo.GetDifficultyCode(*task.DifficultyID)
		if err == nil {
			mult = complexityMultiplier(code)
		}
	}

	basePoints := int(math.Round(10 * mult))

	// 4. On-time bonus: awarded only when a due date is set and not yet passed.
	onTimePts := 0
	if task.DueDate != nil && now.Before(*task.DueDate) {
		onTimePts = 5
	}

	// 5. Quality bonus: awarded immediately; reversed on reopen.
	qualityPts := 3

	// 6. Run all inserts + user_gamification update atomically.
	return s.db.Transaction(func(tx *gorm.DB) error {
		taskID := task.ID

		if pts := bgt.take(basePoints); pts > 0 {
			if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
				UserID:   developerUserID,
				TaskID:   &taskID,
				Points:   pts,
				Reason:   ReasonTaskCompleted,
				EarnedAt: now,
				Metadata: jsonMeta(map[string]interface{}{"multiplier": mult}),
			}); err != nil {
				return err
			}
		}

		if !bgt.exhausted() && onTimePts > 0 {
			if pts := bgt.take(onTimePts); pts > 0 {
				if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
					UserID:   developerUserID,
					TaskID:   &taskID,
					Points:   pts,
					Reason:   ReasonOnTimeBonus,
					EarnedAt: now,
				}); err != nil {
					return err
				}
			}
		}

		if !bgt.exhausted() {
			if pts := bgt.take(qualityPts); pts > 0 {
				if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
					UserID:   developerUserID,
					TaskID:   &taskID,
					Points:   pts,
					Reason:   ReasonQualityBonus,
					EarnedAt: now,
				}); err != nil {
					return err
				}
			}
		}

		// 7. Combo bonus: award +5 when this is exactly the 3rd completion today.
		if !bgt.exhausted() {
			count, err := s.repo.CountCompletionsToday(tx, developerUserID)
			if err != nil {
				return err
			}
			if count == 3 {
				if pts := bgt.take(5); pts > 0 {
					if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
						UserID:   developerUserID,
						Points:   pts,
						Reason:   ReasonComboBonus,
						EarnedAt: now,
					}); err != nil {
						return err
					}
				}
			}
		}

		// 8. Streak update and streak bonuses.
		ug, err := s.repo.GetOrCreateUserGamification(developerUserID)
		if err != nil {
			return err
		}

		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		if ug.LastActiveDate == nil || ug.LastActiveDate.Before(today) {
			prev := ug.LastActiveDate
			ug.LastActiveDate = &today

			yesterday := today.AddDate(0, 0, -1)
			if prev != nil && prev.Equal(yesterday) {
				ug.CurrentStreak++
			} else {
				ug.CurrentStreak = 1
			}
			if ug.CurrentStreak > ug.LongestStreak {
				ug.LongestStreak = ug.CurrentStreak
			}

			// +2/day streak bonus after 3 consecutive active days.
			if ug.CurrentStreak > 3 && !bgt.exhausted() {
				if pts := bgt.take(2); pts > 0 {
					if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
						UserID:   developerUserID,
						Points:   pts,
						Reason:   ReasonStreakBonus,
						EarnedAt: now,
						Metadata: jsonMeta(map[string]interface{}{"streak": ug.CurrentStreak}),
					}); err != nil {
						return err
					}
				}
			}

			// +10 milestone every 5-day streak.
			if ug.CurrentStreak > 0 && ug.CurrentStreak%5 == 0 && !bgt.exhausted() {
				if pts := bgt.take(10); pts > 0 {
					if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
						UserID:   developerUserID,
						Points:   pts,
						Reason:   ReasonStreakMilestone,
						EarnedAt: now,
						Metadata: jsonMeta(map[string]interface{}{"streak": ug.CurrentStreak}),
					}); err != nil {
						return err
					}
				}
			}
		}

		// 9. Recompute total from all transactions (includes newly inserted rows).
		newTotal, err := s.repo.SumAllPointsForUser(tx, developerUserID)
		if err != nil {
			return err
		}
		ug.TotalPoints = newTotal
		ug.CurrentLevel = ComputeLevel(newTotal)
		ug.UpdatedAt = now

		return s.repo.SaveUserGamification(tx, ug)
	})
}

// ─── OnTaskReopened ───────────────────────────────────────────────────────────

// OnTaskReopened applies a rework penalty by reversing any on_time_bonus and
// quality_bonus that were previously awarded for this task and not yet reversed.
func (s *Service) OnTaskReopened(developerUserID int64, task *models.Task) error {
	if developerUserID == 0 {
		return nil
	}

	taskID := task.ID
	now := time.Now().UTC()

	// Find previously awarded bonuses.
	bonuses, err := s.repo.GetTransactionsByTaskAndReason(taskID, developerUserID,
		[]string{ReasonOnTimeBonus, ReasonQualityBonus})
	if err != nil {
		return err
	}
	if len(bonuses) == 0 {
		return nil // nothing to reverse
	}

	// Total awarded so far.
	awardedSum := 0
	for _, b := range bonuses {
		awardedSum += b.Points
	}

	// Total already penalized (handles repeated reopen/complete cycles).
	alreadyPenalized, err := s.repo.SumReworkPenalties(taskID, developerUserID)
	if err != nil {
		return err
	}

	netToReverse := awardedSum - alreadyPenalized
	if netToReverse <= 0 {
		return nil // already fully reversed
	}

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
			UserID:   developerUserID,
			TaskID:   &taskID,
			Points:   -netToReverse,
			Reason:   ReasonReworkPenalty,
			EarnedAt: now,
		}); err != nil {
			return err
		}

		ug, err := s.repo.GetOrCreateUserGamification(developerUserID)
		if err != nil {
			return err
		}

		newTotal, err := s.repo.SumAllPointsForUser(tx, developerUserID)
		if err != nil {
			return err
		}
		if newTotal < 0 {
			newTotal = 0
		}
		ug.TotalPoints = newTotal
		ug.CurrentLevel = ComputeLevel(newTotal)
		ug.UpdatedAt = now

		return s.repo.SaveUserGamification(tx, ug)
	})
}

// ─── GiveKudos ────────────────────────────────────────────────────────────────

// GiveKudos records a kudos and awards +5 pts to the recipient.
// Enforces the server-side limit of MaxKudosPerWeek per giver per week.
func (s *Service) GiveKudos(fromUserID, toUserID int64, taskID *uint, message string) error {
	if fromUserID == toUserID {
		return errors.New("cannot give kudos to yourself")
	}

	given, err := s.repo.CountKudosGivenThisWeek(fromUserID)
	if err != nil {
		return err
	}
	if given >= MaxKudosPerWeek {
		return errors.New("weekly kudos limit reached (max 3 per week)")
	}

	now := time.Now().UTC()

	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.repo.InsertKudos(&models.Kudos{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
			TaskID:     taskID,
			Message:    message,
			CreatedAt:  now,
		}); err != nil {
			return err
		}

		// Check daily cap for recipient.
		todayPts, err := s.repo.SumPointsToday(toUserID)
		if err != nil {
			return err
		}
		if todayPts >= DailyCap {
			return nil // recipient is at cap; kudos recorded but no points
		}

		pts := 5
		remaining := DailyCap - todayPts
		if pts > remaining {
			pts = remaining
		}

		if err := s.repo.InsertTransaction(tx, &models.PointTransaction{
			UserID:   toUserID,
			TaskID:   taskID,
			Points:   pts,
			Reason:   ReasonKudosReceived,
			EarnedAt: now,
			Metadata: jsonMeta(map[string]interface{}{"from_user_id": fromUserID}),
		}); err != nil {
			return err
		}

		ug, err := s.repo.GetOrCreateUserGamification(toUserID)
		if err != nil {
			return err
		}
		newTotal, err := s.repo.SumAllPointsForUser(tx, toUserID)
		if err != nil {
			return err
		}
		ug.TotalPoints = newTotal
		ug.CurrentLevel = ComputeLevel(newTotal)
		ug.UpdatedAt = now

		return s.repo.SaveUserGamification(tx, ug)
	})
}

// ─── Leaderboard / stats ──────────────────────────────────────────────────────

func (s *Service) GetLeaderboard(limit int) ([]LeaderboardEntry, error) {
	rows, err := s.repo.GetLeaderboard(limit)
	if err != nil {
		return nil, err
	}
	result := make([]LeaderboardEntry, len(rows))
	for i, r := range rows {
		lvl := ComputeLevel(r.TotalPoints)
		lvlName := levelName(lvl)
		result[i] = LeaderboardEntry{
			Rank:          i + 1,
			UserID:        r.UserID,
			FullName:      r.FullName,
			Photo:         r.Photo,
			RollingPoints: r.RollingPoints,
			TotalPoints:   r.TotalPoints,
			CurrentLevel:  lvl,
			LevelName:     lvlName,
			CurrentStreak: r.CurrentStreak,
		}
	}
	return result, nil
}

func (s *Service) GetUserStats(userID int64) (*UserStatsResponse, error) {
	ug, err := s.repo.GetOrCreateUserGamification(userID)
	if err != nil {
		return nil, err
	}
	rolling, err := s.repo.GetRolling30dPoints(userID)
	if err != nil {
		return nil, err
	}
	todayPts, err := s.repo.SumPointsToday(userID)
	if err != nil {
		return nil, err
	}
	return &UserStatsResponse{
		UserID:        userID,
		TotalPoints:   ug.TotalPoints,
		Rolling30d:    rolling,
		CurrentLevel:  ug.CurrentLevel,
		LevelName:     levelName(ug.CurrentLevel),
		CurrentStreak: ug.CurrentStreak,
		LongestStreak: ug.LongestStreak,
		TodayPoints:   todayPts,
		DailyCap:      DailyCap,
		DailyCapHit:   todayPts >= DailyCap,
		Levels:        Levels,
	}, nil
}

func (s *Service) GetPointsHistory(userID int64, limit int) ([]PointTransactionResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	pts, err := s.repo.GetPointsHistory(userID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]PointTransactionResponse, len(pts))
	for i, p := range pts {
		result[i] = PointTransactionResponse{
			ID:       p.ID,
			TaskID:   p.TaskID,
			Points:   p.Points,
			Reason:   p.Reason,
			EarnedAt: p.EarnedAt,
		}
	}
	return result, nil
}

func (s *Service) GetKudosStatus(fromUserID int64) (*KudosStatusResponse, error) {
	given, err := s.repo.CountKudosGivenThisWeek(fromUserID)
	if err != nil {
		return nil, err
	}
	return &KudosStatusResponse{
		GivenThisWeek: int(given),
		MaxPerWeek:    MaxKudosPerWeek,
	}, nil
}

func levelName(level int) string {
	for _, l := range Levels {
		if l.Level == level {
			return l.Name
		}
	}
	return "Novice"
}
