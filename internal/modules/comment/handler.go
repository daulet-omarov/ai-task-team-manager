package comment

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
// @Summary Add comment
// @Description Add a comment to a task; caller must be a board member
// @Tags Comment
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param taskId path int true "Task ID"
// @Param request body CreateCommentRequest true "Comment body"
// @Success 201 {object} CommentResponse
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId}/comments [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	var req CreateCommentRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	c, err := h.service.Create(uint(taskID), userID, req)
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

	response.JSON(w, http.StatusCreated, c)
}

// GetByTaskID godoc
// @Summary Get task comments
// @Description Returns all comments for a task; caller must be a board member
// @Tags Comment
// @Security BearerAuth
// @Produce json
// @Param taskId path int true "Task ID"
// @Success 200 {array} CommentResponse
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId}/comments [get]
func (h *Handler) GetByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	userID := middleware.GetUserID(r)

	comments, err := h.service.GetByTaskID(uint(taskID), userID)
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

	response.JSON(w, http.StatusOK, comments)
}

// Delete godoc
// @Summary Delete comment
// @Description Delete a comment; only the author can delete
// @Tags Comment
// @Security BearerAuth
// @Param commentId path int true "Comment ID"
// @Success 200 {string} string "deleted"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "comment not found"
// @Router /comments/{commentId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.ParseUint(chi.URLParam(r, "commentId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid comment id")
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.Delete(uint(commentID), userID); err != nil {
		switch err.Error() {
		case "access denied":
			response.Error(w, http.StatusForbidden, err.Error())
		case "comment not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
