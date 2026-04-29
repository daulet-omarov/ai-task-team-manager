package employee

import (
	"errors"
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(userID uint, req CreateEmployeeRequest) error {
	birthday, err := time.Parse("2006-01-02", req.Birthday)
	if err != nil {
		return errors.New("invalid birthday format, use YYYY-MM-DD")
	}

	e := &models.Employee{
		ID:          userID, // employee.id == user.id always (1-to-1)
		UserID:      userID,
		FullName:    req.FullName,
		Photo:       req.Photo,
		Email:       req.Email,
		Birthday:    birthday,
		PhoneNumber: req.PhoneNumber,
		GenderID:    req.GenderID,
	}

	return s.repo.Create(e)
}

func (s *Service) GetByID(id uint) (*EmployeeResponse, error) {
	e, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("employee not found")
	}
	return toResponse(e), nil
}

func (s *Service) GetByUserID(userID uint) (*EmployeeResponse, error) {
	e, err := s.repo.GetByUserID(userID)
	if err != nil {
		return nil, errors.New("employee not found")
	}
	return toResponse(e), nil
}

func (s *Service) GetAll() ([]*EmployeeResponse, error) {
	employees, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	var result []*EmployeeResponse
	for _, e := range employees {
		result = append(result, toResponse(e))
	}
	return result, nil
}

func (s *Service) Update(userID uint, req UpdateEmployeeRequest) error {
	e, err := s.repo.GetByUserID(userID)
	if err != nil {
		return errors.New("employee not found")
	}

	if req.FullName != "" {
		e.FullName = req.FullName
	}
	if req.Photo != "" {
		e.Photo = req.Photo
	}
	if req.Email != "" {
		e.Email = req.Email
	}
	if req.GenderID != 0 {
		e.GenderID = req.GenderID
	}
	if req.PhoneNumber != "" {
		e.PhoneNumber = req.PhoneNumber
	}
	if req.Birthday != "" {
		birthday, err := time.Parse("2006-01-02", req.Birthday)
		if err != nil {
			return errors.New("invalid birthday format, use YYYY-MM-DD")
		}
		e.Birthday = birthday
	}

	return s.repo.Update(e)
}

func (s *Service) Delete(userID uint) error {
	return s.repo.Delete(userID)
}

// GetProfile returns profile info + activity dashboard for a given employee/user ID.
// Since employee.id == user.id (1-to-1), both are interchangeable.
func (s *Service) GetProfile(id uint) (*ProfileResponse, error) {
	profile, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	activities, err := s.GetActivities(id)
	if err != nil {
		return nil, err
	}
	return &ProfileResponse{
		Profile:    profile,
		Activities: activities,
	}, nil
}

func (s *Service) Exists(userID uint) (bool, error) {
	_, err := s.repo.GetByUserID(userID)
	if err != nil {
		return false, nil
	}
	return true, nil
}

const contributionStaleness = time.Hour

func (s *Service) GetActivities(employeeID uint) (*ActivitiesResponse, error) {
	stored, lastComputed, err := s.repo.getStoredActivities(employeeID)
	if err != nil {
		return nil, err
	}

	// First request ever: full recompute across all history.
	if len(stored) == 0 {
		rows, err := s.repo.computeActivities(employeeID)
		if err != nil {
			return nil, err
		}
		if err := s.repo.upsertActivities(employeeID, rows); err != nil {
			return nil, err
		}
		return buildActivitiesResponse(rows), nil
	}

	// Data is fresh — return as-is.
	if time.Since(lastComputed) < contributionStaleness {
		return buildActivitiesResponse(stored), nil
	}

	// Stale: find the earliest date that could be affected by changes since
	// the last computation. This avoids recomputing untouched history.
	minDate, hasChanges, err := s.repo.findMinAffectedDate(employeeID, lastComputed)
	if err != nil {
		return nil, err
	}

	// Nothing changed at all — just refresh timestamps and return stored data.
	if !hasChanges {
		_ = s.repo.touchActivities(employeeID)
		return buildActivitiesResponse(stored), nil
	}

	// Recompute only from the earliest affected date onwards.
	recent, err := s.repo.computeActivitiesSince(employeeID, minDate)
	if err != nil {
		return nil, err
	}
	if err := s.repo.upsertActivitiesSince(employeeID, minDate, recent); err != nil {
		return nil, err
	}

	// Return the full picture: untouched old history + refreshed window.
	all, _, err := s.repo.getStoredActivities(employeeID)
	if err != nil {
		return nil, err
	}
	return buildActivitiesResponse(all), nil
}

func buildActivitiesResponse(rows []dailyActivityRow) *ActivitiesResponse {
	activities := make([]DailyContribution, 0, len(rows))
	total := 0
	for _, r := range rows {
		activities = append(activities, DailyContribution{Date: r.Date, Count: r.Count})
		total += r.Count
	}

	return &ActivitiesResponse{
		Activities:         activities,
		TotalContributions: total,
		TotalActiveDays:    len(rows),
		MaxStreak:          calcMaxStreak(rows),
		CurrentStreak:      calcCurrentStreak(rows),
	}
}

// calcCurrentStreak counts consecutive active days going backwards from today.
// Yesterday counts too — the user may not have contributed yet today.
func calcCurrentStreak(rows []dailyActivityRow) int {
	if len(rows) == 0 {
		return 0
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	last, _ := time.Parse("2006-01-02", rows[len(rows)-1].Date)
	last = last.UTC()

	// Streak is broken if the last active day is not today or yesterday.
	if last.Before(yesterday) {
		return 0
	}

	streak := 1
	for i := len(rows) - 2; i >= 0; i-- {
		curr, _ := time.Parse("2006-01-02", rows[i].Date)
		next, _ := time.Parse("2006-01-02", rows[i+1].Date)
		if next.Sub(curr) == 24*time.Hour {
			streak++
		} else {
			break
		}
	}
	return streak
}

func calcMaxStreak(rows []dailyActivityRow) int {
	if len(rows) == 0 {
		return 0
	}
	max, cur := 1, 1
	for i := 1; i < len(rows); i++ {
		prev, _ := time.Parse("2006-01-02", rows[i-1].Date)
		curr, _ := time.Parse("2006-01-02", rows[i].Date)
		if curr.Sub(prev) == 24*time.Hour {
			cur++
			if cur > max {
				max = cur
			}
		} else {
			cur = 1
		}
	}
	return max
}

func (s *Service) GetAchievements(employeeID uint) ([]AchievementResponse, error) {
	counts, err := s.repo.getAchievementCounts(employeeID)
	if err != nil {
		return nil, err
	}
	dates, err := s.repo.getActivityDates(employeeID)
	if err != nil {
		return nil, err
	}
	return computeAchievements(counts, dates), nil
}

func computeAchievements(c *achievementCountRow, dates []string) []AchievementResponse {
	maxStreak := calcMaxStreakFromDates(dates)
	totalActiveDays := len(dates)
	streakCount := countStreaks(dates)

	lvl := func(val, l1, l2, l3 int64) int {
		switch {
		case val >= l3:
			return 3
		case val >= l2:
			return 2
		case val >= l1:
			return 1
		default:
			return 0
		}
	}

	results := []AchievementResponse{
		{"closer", lvl(c.Closer, 10, 50, 200)},
		{"finisher", 0}, // no due-date field in tasks table
		{"challenger", lvl(c.Challenger, 10, 50, 200)},
		{"elite", lvl(c.Elite, 5, 25, 100)},
		{"perfectionist", lvl(c.Perfectionist, 5, 20, 50)},
		{"cleanworker", lvl(c.Cleanworker, 10, 50, 150)},
		{"consistency", lvl(int64(maxStreak), 3, 7, 11)},
		{"unstoppable", lvl(int64(totalActiveDays), 50, 100, 200)},
		{"onfire", lvl(int64(streakCount), 3, 7, 30)},
		{"teamplayer", lvl(c.Teamplayer, 5, 15, 50)},
		{"reviewer", lvl(c.Reviewer, 10, 50, 150)},
		{"communicator", lvl(c.Communicator, 50, 200, 1000)},
		{"pollmaster", lvl(c.Pollmaster, 10, 100, 1000)},
		{"influencer", lvl(c.Influencer, 50, 300, 1000)},
		{"voicepioneer", lvl(c.Voicepioneer, 20, 100, 1000)},
		{"broadcaster", lvl(c.Broadcaster, 20, 100, 1000)},
	}

	level3Count, anyLevelCount := 0, 0
	for _, a := range results {
		if a.Level >= 3 {
			level3Count++
		}
		if a.Level >= 1 {
			anyLevelCount++
		}
	}

	prestigeLevel := func(unlocked bool) int {
		if unlocked {
			return 4
		}
		return 0
	}

	results = append(results,
		AchievementResponse{"legend", prestigeLevel(level3Count >= 10)},
		AchievementResponse{"master", prestigeLevel(anyLevelCount >= 25)},
		AchievementResponse{"grandmaster", prestigeLevel(anyLevelCount >= 50)},
	)

	return results
}

func calcMaxStreakFromDates(dates []string) int {
	if len(dates) == 0 {
		return 0
	}
	max, cur := 1, 1
	for i := 1; i < len(dates); i++ {
		prev, _ := time.Parse("2006-01-02", dates[i-1])
		curr, _ := time.Parse("2006-01-02", dates[i])
		if curr.Sub(prev) == 24*time.Hour {
			cur++
			if cur > max {
				max = cur
			}
		} else {
			cur = 1
		}
	}
	return max
}

func countStreaks(dates []string) int {
	if len(dates) == 0 {
		return 0
	}
	count := 1
	for i := 1; i < len(dates); i++ {
		prev, _ := time.Parse("2006-01-02", dates[i-1])
		curr, _ := time.Parse("2006-01-02", dates[i])
		if curr.Sub(prev) > 24*time.Hour {
			count++
		}
	}
	return count
}

// --- helper ---

func toResponse(e *models.Employee) *EmployeeResponse {
	return &EmployeeResponse{
		ID:          e.ID,
		UserID:      e.UserID,
		FullName:    e.FullName,
		Photo:       e.Photo,
		Email:       e.Email,
		PhoneNumber: e.PhoneNumber,
		Birthday:    e.Birthday.Format("2006-01-02"),
		Gender:      e.Gender,
	}
}
