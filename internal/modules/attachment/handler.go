package attachment

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Upload godoc
// @Summary Upload attachment
// @Description Upload a file attachment to a task; caller must be a board member
// @Tags Attachment
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param taskId path int true "Task ID"
// @Param file formData file true "File to upload"
// @Success 201 {object} AttachmentResponse
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId}/attachments [post]
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, uploader.MaxAttachmentSize)
	if err := r.ParseMultipartForm(uploader.MaxAttachmentSize); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid form data or file too large")
		return
	}

	_, fh, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "file is required")
		return
	}

	userID := middleware.GetUserID(r)

	a, err := h.service.Upload(uint(taskID), userID, fh, r)
	if err != nil {
		switch err.Error() {
		case "access denied":
			response.Error(w, http.StatusForbidden, err.Error())
		case "task not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.JSON(w, http.StatusCreated, a)
}

// GetByTaskID godoc
// @Summary Get task attachments
// @Description Returns all attachments for a task; caller must be a board member
// @Tags Attachment
// @Security BearerAuth
// @Produce json
// @Param taskId path int true "Task ID"
// @Success 200 {array} AttachmentResponse
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "task not found"
// @Router /tasks/{taskId}/attachments [get]
func (h *Handler) GetByTaskID(w http.ResponseWriter, r *http.Request) {
	taskID, err := strconv.ParseUint(chi.URLParam(r, "taskId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid task id")
		return
	}

	userID := middleware.GetUserID(r)

	attachments, err := h.service.GetByTaskID(uint(taskID), userID, r)
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

	response.JSON(w, http.StatusOK, attachments)
}

// Delete godoc
// @Summary Delete attachment
// @Description Delete an attachment; caller must be a board member
// @Tags Attachment
// @Security BearerAuth
// @Param attachmentId path int true "Attachment ID"
// @Success 200 {string} string "deleted"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "attachment not found"
// @Router /attachments/{attachmentId} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	attachmentID, err := strconv.ParseUint(chi.URLParam(r, "attachmentId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid attachment id")
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.Delete(uint(attachmentID), userID); err != nil {
		switch err.Error() {
		case "access denied":
			response.Error(w, http.StatusForbidden, err.Error())
		case "attachment not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
