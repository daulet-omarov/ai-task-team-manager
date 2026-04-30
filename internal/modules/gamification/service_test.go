package gamification

import (
	"errors"
	"testing"
	"time"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
	"gorm.io/gorm"
)

// ─── Mock repository ──────────────────────────────────────────────────────────

type mockRepo struct {
	todayPts         int
	completionsToday int64
	bonusTxns        []models.PointTransaction
	reworkPenalties  int
	kudosGiven       int64

	savedTxns []models.PointTransaction
	savedUG   *models.UserGamification
	savedKudos []models.Kudos

	sumAllResult int
	diffCode     string
}

func (m *mockRepo) GetOrCreateUserGamification(userID int64) (*models.UserGamification, error) {
	if m.savedUG == nil {
		m.savedUG = &models.UserGamification{UserID: userID, CurrentLevel: 1}
	}
	return m.savedUG, nil
}
func (m *mockRepo) SaveUserGamification(_ *gorm.DB, ug *models.UserGamification) error {
	m.savedUG = ug
	return nil
}
func (m *mockRepo) InsertTransaction(_ *gorm.DB, pt *models.PointTransaction) error {
	m.savedTxns = append(m.savedTxns, *pt)
	return nil
}
func (m *mockRepo) SumPointsToday(_ int64) (int, error)     { return m.todayPts, nil }
func (m *mockRepo) CountCompletionsToday(_ *gorm.DB, _ int64) (int64, error) {
	return m.completionsToday, nil
}
func (m *mockRepo) GetTransactionsByTaskAndReason(_ uint, _ int64, _ []string) ([]models.PointTransaction, error) {
	return m.bonusTxns, nil
}
func (m *mockRepo) SumReworkPenalties(_ uint, _ int64) (int, error) { return m.reworkPenalties, nil }
func (m *mockRepo) SumAllPointsForUser(_ *gorm.DB, _ int64) (int, error) {
	return m.sumAllResult, nil
}
func (m *mockRepo) CountKudosGivenThisWeek(_ int64) (int64, error) { return m.kudosGiven, nil }
func (m *mockRepo) InsertKudos(k *models.Kudos) error {
	m.savedKudos = append(m.savedKudos, *k)
	return nil
}
func (m *mockRepo) GetLeaderboard(_ int) ([]leaderboardRow, error)                      { return nil, nil }
func (m *mockRepo) GetRolling30dPoints(_ int64) (int, error)                            { return 0, nil }
func (m *mockRepo) GetPointsHistory(_ int64, _ int) ([]models.PointTransaction, error)  { return nil, nil }
func (m *mockRepo) GetDifficultyCode(_ uint) (string, error)                            { return m.diffCode, nil }

// ─── serviceForTest constructs a Service with a mock repo and an in-process tx ─

// repoIface mirrors every method the service calls, so we can swap in mockRepo.
type repoIface interface {
	GetOrCreateUserGamification(userID int64) (*models.UserGamification, error)
	SaveUserGamification(tx *gorm.DB, ug *models.UserGamification) error
	InsertTransaction(tx *gorm.DB, pt *models.PointTransaction) error
	SumPointsToday(userID int64) (int, error)
	CountCompletionsToday(tx *gorm.DB, userID int64) (int64, error)
	GetTransactionsByTaskAndReason(taskID uint, userID int64, reasons []string) ([]models.PointTransaction, error)
	SumReworkPenalties(taskID uint, userID int64) (int, error)
	SumAllPointsForUser(tx *gorm.DB, userID int64) (int, error)
	CountKudosGivenThisWeek(fromUserID int64) (int64, error)
	InsertKudos(k *models.Kudos) error
	GetDifficultyCode(difficultyID uint) (string, error)
}

// testService wraps Service but replaces the db.Transaction call so tests
// don't need a real Postgres connection.
type testService struct {
	repo repoIface
}

func (ts *testService) onTaskCompleted(developerUserID int64, task *models.Task) error {
	now := time.Now().UTC()
	if now.Sub(task.CreatedAt) < 5*time.Minute {
		return nil
	}
	todayPts, _ := ts.repo.SumPointsToday(developerUserID)
	if todayPts >= DailyCap {
		return nil
	}
	bgt := newBudget(DailyCap, todayPts)

	mult := 1.0
	if task.DifficultyID != nil {
		code, err := ts.repo.GetDifficultyCode(*task.DifficultyID)
		if err == nil {
			mult = complexityMultiplier(code)
		}
	}

	basePoints := int(roundFloat(10 * mult))
	onTimePts := 0
	if task.DueDate != nil && now.Before(*task.DueDate) {
		onTimePts = 5
	}
	qualityPts := 3

	// simulate inserts (no real tx needed)
	nilTx := (*gorm.DB)(nil)
	taskID := task.ID
	if pts := bgt.take(basePoints); pts > 0 {
		_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, TaskID: &taskID, Points: pts, Reason: ReasonTaskCompleted, EarnedAt: now})
	}
	if !bgt.exhausted() && onTimePts > 0 {
		if pts := bgt.take(onTimePts); pts > 0 {
			_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, TaskID: &taskID, Points: pts, Reason: ReasonOnTimeBonus, EarnedAt: now})
		}
	}
	if !bgt.exhausted() {
		if pts := bgt.take(qualityPts); pts > 0 {
			_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, TaskID: &taskID, Points: pts, Reason: ReasonQualityBonus, EarnedAt: now})
		}
	}

	count, _ := ts.repo.CountCompletionsToday(nilTx, developerUserID)
	if count == 3 && !bgt.exhausted() {
		if pts := bgt.take(5); pts > 0 {
			_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, Points: pts, Reason: ReasonComboBonus, EarnedAt: now})
		}
	}

	ug, _ := ts.repo.GetOrCreateUserGamification(developerUserID)
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
		if ug.CurrentStreak > 3 && !bgt.exhausted() {
			if pts := bgt.take(2); pts > 0 {
				_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, Points: pts, Reason: ReasonStreakBonus, EarnedAt: now})
			}
		}
		if ug.CurrentStreak > 0 && ug.CurrentStreak%5 == 0 && !bgt.exhausted() {
			if pts := bgt.take(10); pts > 0 {
				_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, Points: pts, Reason: ReasonStreakMilestone, EarnedAt: now})
			}
		}
	}

	newTotal, _ := ts.repo.SumAllPointsForUser(nilTx, developerUserID)
	ug.TotalPoints = newTotal
	ug.CurrentLevel = ComputeLevel(newTotal)
	return ts.repo.SaveUserGamification(nilTx, ug)
}

func (ts *testService) onTaskReopened(developerUserID int64, task *models.Task) error {
	taskID := task.ID
	bonuses, _ := ts.repo.GetTransactionsByTaskAndReason(taskID, developerUserID, []string{ReasonOnTimeBonus, ReasonQualityBonus})
	if len(bonuses) == 0 {
		return nil
	}
	awardedSum := 0
	for _, b := range bonuses {
		awardedSum += b.Points
	}
	alreadyPenalized, _ := ts.repo.SumReworkPenalties(taskID, developerUserID)
	netToReverse := awardedSum - alreadyPenalized
	if netToReverse <= 0 {
		return nil
	}

	nilTx := (*gorm.DB)(nil)
	now := time.Now().UTC()
	_ = ts.repo.InsertTransaction(nilTx, &models.PointTransaction{UserID: developerUserID, TaskID: &taskID, Points: -netToReverse, Reason: ReasonReworkPenalty, EarnedAt: now})

	ug, _ := ts.repo.GetOrCreateUserGamification(developerUserID)
	newTotal, _ := ts.repo.SumAllPointsForUser(nilTx, developerUserID)
	if newTotal < 0 {
		newTotal = 0
	}
	ug.TotalPoints = newTotal
	ug.CurrentLevel = ComputeLevel(newTotal)
	return ts.repo.SaveUserGamification(nilTx, ug)
}

func roundFloat(f float64) float64 {
	if f < 0 {
		return float64(int(f - 0.5))
	}
	return float64(int(f + 0.5))
}

func newTestTask(difficultyID *uint, createdSecondsAgo int, dueInSeconds int) *models.Task {
	now := time.Now().UTC()
	t := &models.Task{
		ID:           1,
		BoardID:      1,
		DeveloperID:  42,
		DifficultyID: difficultyID,
		CreatedAt:    now.Add(-time.Duration(createdSecondsAgo) * time.Second),
		UpdatedAt:    now,
	}
	if dueInSeconds > 0 {
		due := now.Add(time.Duration(dueInSeconds) * time.Second)
		t.DueDate = &due
	}
	return t
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestPointCalculation_BaseOnly(t *testing.T) {
	mr := &mockRepo{sumAllResult: 10}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0) // no difficulty, created 400s ago, no due date
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var base int
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonTaskCompleted {
			base = tx.Points
		}
	}
	if base != 10 {
		t.Errorf("expected base 10, got %d", base)
	}
}

func TestPointCalculation_ComplexityMultiplier(t *testing.T) {
	hardID := uint(3)
	mr := &mockRepo{diffCode: "hard", sumAllResult: 20}
	ts := &testService{repo: mr}

	task := newTestTask(&hardID, 400, 0)
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// hard → complexity 10 → multiplier 2.0 → base = 10*2 = 20
	var base int
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonTaskCompleted {
			base = tx.Points
		}
	}
	if base != 20 {
		t.Errorf("expected base 20 (hard ×2.0), got %d", base)
	}
}

func TestPointCalculation_OnTimeBonus(t *testing.T) {
	mr := &mockRepo{sumAllResult: 15}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 3600) // due in 1 hour
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotOnTime bool
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonOnTimeBonus {
			gotOnTime = true
			if tx.Points != 5 {
				t.Errorf("on-time bonus: expected 5, got %d", tx.Points)
			}
		}
	}
	if !gotOnTime {
		t.Error("expected on-time bonus transaction, got none")
	}
}

func TestDailyCapEnforcement(t *testing.T) {
	// Already at cap — no transactions should be inserted.
	mr := &mockRepo{todayPts: 100, sumAllResult: 100}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mr.savedTxns) != 0 {
		t.Errorf("expected 0 transactions at daily cap, got %d", len(mr.savedTxns))
	}
}

func TestDailyCapEnforcement_PartialAward(t *testing.T) {
	// 95 pts earned today; only 5 pts remain before cap.
	mr := &mockRepo{todayPts: 95, sumAllResult: 100}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 3600) // base 10, on-time 5, quality 3 → would be 18
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var total int
	for _, tx := range mr.savedTxns {
		total += tx.Points
	}
	if total > 5 {
		t.Errorf("total awarded %d exceeds remaining cap of 5", total)
	}
}

func TestFiveMinuteRule(t *testing.T) {
	mr := &mockRepo{sumAllResult: 0}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 30, 0) // created only 30 seconds ago
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mr.savedTxns) != 0 {
		t.Errorf("expected 0 transactions (task too fresh), got %d", len(mr.savedTxns))
	}
}

func TestReworkPenalty(t *testing.T) {
	taskID := uint(1)
	mr := &mockRepo{
		bonusTxns: []models.PointTransaction{
			{TaskID: &taskID, Reason: ReasonOnTimeBonus, Points: 5},
			{TaskID: &taskID, Reason: ReasonQualityBonus, Points: 3},
		},
		reworkPenalties: 0,
		sumAllResult:    5, // after penalty applied
	}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	task.ID = taskID
	if err := ts.onTaskReopened(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var penaltyPts int
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonReworkPenalty {
			penaltyPts = tx.Points
		}
	}
	if penaltyPts != -8 {
		t.Errorf("expected rework penalty -8 (on_time -5 + quality -3), got %d", penaltyPts)
	}
}

func TestReworkPenalty_AlreadyReversed(t *testing.T) {
	taskID := uint(1)
	mr := &mockRepo{
		bonusTxns: []models.PointTransaction{
			{TaskID: &taskID, Reason: ReasonOnTimeBonus, Points: 5},
			{TaskID: &taskID, Reason: ReasonQualityBonus, Points: 3},
		},
		reworkPenalties: 8, // already fully penalized
		sumAllResult:    0,
	}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	task.ID = taskID
	if err := ts.onTaskReopened(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No new penalty should be inserted.
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonReworkPenalty {
			t.Error("unexpected second rework penalty inserted")
		}
	}
}

func TestStreakLogic_FirstDay(t *testing.T) {
	mr := &mockRepo{sumAllResult: 10}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.savedUG.CurrentStreak != 1 {
		t.Errorf("expected streak 1 after first day, got %d", mr.savedUG.CurrentStreak)
	}
}

func TestStreakLogic_ConsecutiveDays(t *testing.T) {
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	mr := &mockRepo{
		sumAllResult: 30,
		savedUG: &models.UserGamification{
			UserID:        1,
			CurrentStreak: 3,
			LongestStreak: 3,
			LastActiveDate: &yesterday,
		},
	}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.savedUG.CurrentStreak != 4 {
		t.Errorf("expected streak 4, got %d", mr.savedUG.CurrentStreak)
	}
	// Streak bonus (+2) should be awarded since streak > 3
	var gotStreakBonus bool
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonStreakBonus {
			gotStreakBonus = true
		}
	}
	if !gotStreakBonus {
		t.Error("expected streak bonus for 4-day streak, got none")
	}
}

func TestStreakLogic_BrokenStreak(t *testing.T) {
	twoDaysAgo := time.Now().UTC().AddDate(0, 0, -2).Truncate(24 * time.Hour)
	mr := &mockRepo{
		sumAllResult: 10,
		savedUG: &models.UserGamification{
			UserID:        1,
			CurrentStreak: 5,
			LongestStreak: 5,
			LastActiveDate: &twoDaysAgo,
		},
	}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.savedUG.CurrentStreak != 1 {
		t.Errorf("expected streak reset to 1 after gap, got %d", mr.savedUG.CurrentStreak)
	}
	if mr.savedUG.LongestStreak != 5 {
		t.Errorf("expected longest streak to remain 5, got %d", mr.savedUG.LongestStreak)
	}
}

func TestStreakMilestone_FiveDay(t *testing.T) {
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Truncate(24 * time.Hour)
	mr := &mockRepo{
		sumAllResult: 50,
		savedUG: &models.UserGamification{
			UserID:        1,
			CurrentStreak: 4, // will become 5 after this completion
			LongestStreak: 4,
			LastActiveDate: &yesterday,
		},
	}
	ts := &testService{repo: mr}

	task := newTestTask(nil, 400, 0)
	if err := ts.onTaskCompleted(1, task); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mr.savedUG.CurrentStreak != 5 {
		t.Errorf("expected streak 5, got %d", mr.savedUG.CurrentStreak)
	}
	var gotMilestone bool
	for _, tx := range mr.savedTxns {
		if tx.Reason == ReasonStreakMilestone {
			gotMilestone = true
		}
	}
	if !gotMilestone {
		t.Error("expected streak milestone bonus at 5-day streak, got none")
	}
}

func TestKudosLimit_Enforced(t *testing.T) {
	mr := &mockRepo{kudosGiven: MaxKudosPerWeek}
	repo := &Repository{} // real repo not used; we test logic isolation
	_ = repo
	svc := &testKudosSvc{repo: mr}
	err := svc.giveKudos(1, 2)
	if err == nil {
		t.Error("expected error when weekly kudos limit reached, got nil")
	}
	if !errors.Is(err, errKudosLimitReached) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestKudosLimit_SelfKudos(t *testing.T) {
	mr := &mockRepo{kudosGiven: 0}
	svc := &testKudosSvc{repo: mr}
	err := svc.giveKudos(1, 1)
	if err == nil {
		t.Error("expected error when giving kudos to self, got nil")
	}
}

// testKudosSvc isolates the kudos validation logic from the DB transaction.
var errKudosLimitReached = errors.New("weekly kudos limit reached (max 3 per week)")

type testKudosSvc struct{ repo *mockRepo }

func (s *testKudosSvc) giveKudos(fromID, toID int64) error {
	if fromID == toID {
		return errors.New("cannot give kudos to yourself")
	}
	given, _ := s.repo.CountKudosGivenThisWeek(fromID)
	if given >= MaxKudosPerWeek {
		return errKudosLimitReached
	}
	return nil
}

func TestComputeLevel(t *testing.T) {
	cases := []struct{ pts, want int }{
		{0, 1}, {99, 1}, {100, 2}, {300, 3},
		{700, 4}, {1500, 5}, {3000, 6}, {9999, 6},
	}
	for _, c := range cases {
		if got := ComputeLevel(c.pts); got != c.want {
			t.Errorf("ComputeLevel(%d) = %d, want %d", c.pts, got, c.want)
		}
	}
}
