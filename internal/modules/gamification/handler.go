package gamification

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/request"
	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetLeaderboard godoc
// @Summary Gamification leaderboard
// @Description Returns top 50 users ordered by rolling 30-day points
// @Tags Gamification
// @Security BearerAuth
// @Produce json
// @Success 200 {array} LeaderboardEntry
// @Router /leaderboard [get]
func (h *Handler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.GetLeaderboard(50)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, entries)
}

// GetMyStats godoc
// @Summary Current user gamification stats
// @Tags Gamification
// @Security BearerAuth
// @Produce json
// @Success 200 {object} UserStatsResponse
// @Router /gamification/me [get]
func (h *Handler) GetMyStats(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	stats, err := h.service.GetUserStats(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// GetUserStats godoc
// @Summary Gamification stats for a specific user
// @Tags Gamification
// @Security BearerAuth
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {object} UserStatsResponse
// @Router /gamification/users/{userId} [get]
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userId")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid user id")
		return
	}
	stats, err := h.service.GetUserStats(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// GetPointsHistory godoc
// @Summary Current user's recent point transactions
// @Tags Gamification
// @Security BearerAuth
// @Produce json
// @Success 200 {array} PointTransactionResponse
// @Router /gamification/history [get]
func (h *Handler) GetPointsHistory(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	history, err := h.service.GetPointsHistory(userID, 50)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, history)
}

// GetKudosStatus godoc
// @Summary How many kudos the current user has given this week
// @Tags Gamification
// @Security BearerAuth
// @Produce json
// @Success 200 {object} KudosStatusResponse
// @Router /gamification/kudos/status [get]
func (h *Handler) GetKudosStatus(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	status, err := h.service.GetKudosStatus(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, status)
}

// GiveKudos godoc
// @Summary Give kudos to another user
// @Tags Gamification
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body GiveKudosRequest true "Kudos request"
// @Success 201 {object} KudosResponse
// @Failure 400 {string} string "validation error"
// @Failure 422 {string} string "weekly limit reached"
// @Router /kudos [post]
func (h *Handler) GiveKudos(w http.ResponseWriter, r *http.Request) {
	var req GiveKudosRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	fromUserID := middleware.GetUserID(r)

	if err := h.service.GiveKudos(fromUserID, req.ToUserID, req.TaskID, req.Message); err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "weekly kudos limit reached (max 3 per week)" ||
			err.Error() == "cannot give kudos to yourself" {
			status = http.StatusUnprocessableEntity
		}
		response.Error(w, status, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, KudosResponse{
		FromUserID: fromUserID,
		ToUserID:   req.ToUserID,
		TaskID:     req.TaskID,
		Message:    req.Message,
	})
}
