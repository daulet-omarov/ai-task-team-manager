package notion

import (
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/request"
	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Import godoc
// @Summary Import tasks from Notion
// @Description Fetches all pages from a Notion database and imports them as tasks. When board_id is 0 a new board is created using the database title.
// @Tags Notion
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body ImportRequest true "Notion import request"
// @Success 200 {object} ImportResult
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "access denied"
// @Router /notion/import [post]
func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	var req ImportRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	result, err := h.service.Import(userID, req)
	if err != nil {
		if err.Error() == "access denied: not a board member" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}
