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
