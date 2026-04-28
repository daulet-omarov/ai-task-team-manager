package task

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/hub"
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/request"
	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	hub     *hub.Hub
}

func NewHandler(service *Service, h *hub.Hub) *Handler {
	return &Handler{service: service, hub: h}
}

// Create godoc
// @Summary Add task to board
// @Description Create a task inside a board; accepts multipart/form-data so attachments can be uploaded in the same request
// @Tags Task
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param boardId       path     int    true  "Board ID"
// @Param title         formData string true  "Title"
// @Param description   formData string false "Description"
// @Param priority_id   formData int    false "Priority ID"
// @Param difficulty_id formData int    false "Difficulty ID"
// @Param assignee_id   formData int    false "Assignee employee ID"
// @Param tester_id     formData int    false "Tester employee ID"
// @Param attachments   formData file   false "Attachments (repeat field for multiple files)"
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

	r.Body = http.MaxBytesReader(w, r.Body, uploader.MaxAttachmentSize*10) // allow multiple files
	if err := r.ParseMultipartForm(uploader.MaxAttachmentSize * 10); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid form data")
		return
	}

	title := r.FormValue("title")
	if title == "" {
		response.Error(w, http.StatusBadRequest, "title is required")
		return
	}

	req := CreateTaskRequest{
		Title:       title,
		Description: r.FormValue("description"),
		AssigneeID:  parseUintForm(r, "assignee_id"),
		TesterID:    parseUintForm(r, "tester_id"),
		PriorityID:  parseUintForm(r, "priority_id"),
	}

	if v := r.FormValue("difficulty_id"); v != "" {
		if id := parseUintForm(r, "difficulty_id"); id != 0 {
			req.DifficultyID = &id
		}
	}

	if r.MultipartForm != nil {
		req.Files = r.MultipartForm.File["attachments"]
	}

	userID := middleware.GetUserID(r)

	task, err := h.service.Create(uint(boardID), userID, req, r)
	if err != nil {
		if err.Error() == "access denied: not a board member" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.hub.Broadcast(uint(boardID), hub.Event{Type: "task_created", Data: task})
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

	tasks, err := h.service.GetByBoardID(uint(boardID), userID, r)
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

	task, err := h.service.GetByID(uint(taskID), userID, r)
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

// Delete godoc
// @Summary Delete task
// @Description Delete a task by ID; caller must be a board member
// @Tags Task
// @Security BearerAuth
// @Param taskId path int true "Task ID"
// @Success 200 {string} string "deleted"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	userID := middleware.GetUserID(r)

	boardID, err := h.service.Delete(uint(taskID), userID)
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

	h.hub.Broadcast(boardID, hub.Event{Type: "task_deleted", Data: map[string]uint{"id": uint(taskID), "board_id": boardID}})
	w.WriteHeader(http.StatusOK)
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

	task, err := h.service.Update(uint(taskID), userID, req, r)
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

	h.hub.Broadcast(task.BoardID, hub.Event{Type: "task_updated", Data: task})
	response.JSON(w, http.StatusOK, task)
}

func parseUintForm(r *http.Request, field string) uint {
	v := r.FormValue(field)
	if v == "" {
		return 0
	}
	id, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint(id)
}
