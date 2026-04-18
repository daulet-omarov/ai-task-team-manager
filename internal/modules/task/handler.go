package task

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

// Create godoc
// @Summary Add task to board
// @Description Create a task inside a board; caller must be a board member
// @Tags Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param boardId path int true "Board ID"
// @Param request body CreateTaskRequest true "Create task request"
// @Success 201 {object} TaskResponse
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Router /boards/{boardId}/tasks [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.ParseUint(chi.URLParam(r, "boardId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	var req CreateTaskRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	task, err := h.service.Create(uint(boardID), userID, req)
	if err != nil {
		if err.Error() == "access denied: not a board member" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, task)
}

// GetByBoardID godoc
// @Summary Get board tasks
// @Description Returns all tasks for a board; caller must be a member
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Param boardId path int true "Board ID"
// @Success 200 {array} TaskResponse
// @Failure 403 {string} string "access denied"
// @Router /boards/{boardId}/tasks [get]
func (h *Handler) GetByBoardID(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.ParseUint(chi.URLParam(r, "boardId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	userID := middleware.GetUserID(r)

	tasks, err := h.service.GetByBoardID(uint(boardID), userID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, tasks)
}

// GetByID godoc
// @Summary Get task
// @Description Get task by ID; caller must be a board member
// @Tags Task
// @Security BearerAuth
// @Produce json
// @Param taskId path int true "Task ID"
// @Success 200 {object} TaskResponse
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	userID := middleware.GetUserID(r)

	task, err := h.service.GetByID(uint(taskID), userID)
	if err != nil {
		if err.Error() == "access denied" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, task)
}

// Update godoc
// @Summary Update task
// @Description Partially update a task; caller must be a board member
// @Tags Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param taskId path int true "Task ID"
// @Param request body UpdateTaskRequest true "Update task request"
// @Success 200 {object} TaskResponse
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId} [patch]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req UpdateTaskRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	task, err := h.service.Update(uint(taskID), userID, req)
	if err != nil {
		switch err.Error() {
		case "access denied":
			response.Error(w, http.StatusForbidden, err.Error())
		case "task not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.JSON(w, http.StatusOK, task)
}
