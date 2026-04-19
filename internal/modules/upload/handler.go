package upload

import (
	"net/http"

	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	"github.com/daulet-omarov/ai-task-team-manager/pkg/uploader"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// UploadPhoto godoc
// @Summary Upload photo
// @Description Upload a profile photo (max 5 MB, jpeg/png/gif/webp). Returns a URL to use in employee fields.
// @Tags Upload
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param photo formData file true "Photo file"
// @Success 200 {object} map[string]string "url"
// @Failure 400 {string} string "bad request"
// @Router /upload/photo [post]
func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, uploader.MaxFileSize)

	if err := r.ParseMultipartForm(uploader.MaxFileSize); err != nil {
		if err.Error() == "http: request body too large" {
			response.Error(w, http.StatusBadRequest, "file too large (max 5 MB)")
		} else {
			response.Error(w, http.StatusBadRequest, "invalid multipart form")
		}
		return
	}

	_, fh, err := r.FormFile("photo")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "field 'photo' is required")
		return
	}

	path, err := uploader.SavePhoto(fh)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"url": uploader.FullURL(r, path)})
}
