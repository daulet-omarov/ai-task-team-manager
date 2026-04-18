package board

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/request"
	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetDashboard godoc
// @Summary Dashboard
// @Description Returns user's boards and isFirstLogin flag
// @Tags Board
// @Security BearerAuth
// @Produce json
// @Success 200 {object} DashboardResponse
// @Router /dashboard [get]
func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	data, err := h.service.GetDashboard(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, data)
}

// Create godoc
// @Summary Create board
// @Description Create a new board; the caller becomes the owner
// @Tags Board
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateBoardRequest true "Create board request"
// @Success 201 {object} BoardResponse
// @Failure 400 {string} string "bad request"
// @Router /boards [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateBoardRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	board, err := h.service.Create(userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, board)
}

// GetByID godoc
// @Summary Get board
// @Description Get board by ID (user must be a member)
// @Tags Board
// @Security BearerAuth
// @Produce json
// @Param id path int true "Board ID"
// @Success 200 {object} BoardResponse
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "board not found"
// @Router /boards/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	userID := middleware.GetUserID(r)

	board, err := h.service.GetByID(uint(id), userID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, board)
}

// GetStatuses godoc
// @Summary Get board statuses
// @Description Returns statuses of a board ordered by position; caller must be a member
// @Tags Board
// @Security BearerAuth
// @Produce json
// @Param id path int true "Board ID"
// @Success 200 {array} StatusResponse
// @Failure 403 {string} string "access denied"
// @Router /boards/{id}/statuses [get]
func (h *Handler) GetStatuses(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	userID := middleware.GetUserID(r)

	statuses, err := h.service.GetStatuses(uint(id), userID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, statuses)
}

// CreateStatus godoc
// @Summary Create board status
// @Description Add a new status to a board; caller must be a member
// @Tags Board
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Board ID"
// @Param request body CreateStatusRequest true "Create status request"
// @Success 201 {object} StatusResponse
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Router /boards/{id}/statuses [post]
func (h *Handler) CreateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	var req CreateStatusRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	status, err := h.service.CreateStatus(uint(id), userID, req)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, status)
}

// ReorderStatuses godoc
// @Summary Reorder board statuses
// @Description Update the position of statuses on a board; caller must be a member
// @Tags Board
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Board ID"
// @Param request body ReorderStatusesRequest true "New order"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Router /boards/{id}/statuses/reorder [patch]
func (h *Handler) ReorderStatuses(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	var req ReorderStatusesRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.ReorderStatuses(uint(id), userID, req); err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetMembers godoc
// @Summary Get board members
// @Description Returns all members of a board with their employee profile info
// @Tags Board
// @Security BearerAuth
// @Produce json
// @Param id path int true "Board ID"
// @Success 200 {array} MemberResponse
// @Failure 403 {string} string "access denied"
// @Router /boards/{id}/members [get]
func (h *Handler) GetMembers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	userID := middleware.GetUserID(r)

	members, err := h.service.GetMembers(uint(id), userID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, members)
}
