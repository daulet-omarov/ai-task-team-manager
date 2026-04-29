package employee

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
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

// CreateEmployee godoc
// @Summary Create employee
// @Description Create a new employee profile. Accepts multipart/form-data so photo can be uploaded in the same request.
// @Tags Employee
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param photo        formData file   false "Profile photo (jpeg/png/gif/webp, max 5 MB)"
// @Param full_name    formData string true  "Full name"
// @Param email        formData string true  "Email"
// @Param gender_id    formData int    true  "Gender ID"
// @Param birthday     formData string true  "Birthday (YYYY-MM-DD)"
// @Param phone_number formData string false "Phone number"
// @Success 201 {string} string "created"
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "server error"
// @Router /employees [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, uploader.MaxFileSize)
	if err := r.ParseMultipartForm(uploader.MaxFileSize); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid form data")
		return
	}

	// Save photo if provided (optional)
	_, fh, _ := r.FormFile("photo")
	photoPath, err := uploader.SavePhoto(fh) // returns ("", nil) when fh is nil
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	photoURL := uploader.FullURL(r, photoPath)

	genderID, err := parseUintField(r, "gender_id")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "gender_id must be a positive integer")
		return
	}

	req := CreateEmployeeRequest{
		FullName:    r.FormValue("full_name"),
		Email:       r.FormValue("email"),
		Birthday:    r.FormValue("birthday"),
		PhoneNumber: r.FormValue("phone_number"),
		Photo:       photoURL,
		GenderID:    genderID,
	}

	if err := validateCreateRequest(req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := uint(middleware.GetUserID(r))
	if err := h.service.Create(userID, req); err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetEmployee godoc
// @Summary Get my employee profile
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Success 200 {object} EmployeeResponse
// @Failure 404 {string} string "not found"
// @Router /employees/me [get]
func (h *Handler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	userID := uint(middleware.GetUserID(r))

	emp, err := h.service.GetByUserID(userID)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	emp.Photo = uploader.FullURL(r, emp.Photo)
	response.JSON(w, http.StatusOK, emp)
}

// GetAllEmployees godoc
// @Summary Get all employees
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Success 200 {array} EmployeeResponse
// @Failure 500 {string} string "server error"
// @Router /employees [get]
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	employees, err := h.service.GetAll()
	if err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	for _, emp := range employees {
		emp.Photo = uploader.FullURL(r, emp.Photo)
	}
	response.JSON(w, http.StatusOK, employees)
}

// UpdateEmployee godoc
// @Summary Update employee profile
// @Description Accepts multipart/form-data. Send only the fields you want to change.
// @Tags Employee
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param photo        formData file   false "New profile photo"
// @Param full_name    formData string false "Full name"
// @Param email        formData string false "Email"
// @Param gender_id    formData int    false "Gender ID"
// @Param birthday     formData string false "Birthday (YYYY-MM-DD)"
// @Param phone_number formData string false "Phone number"
// @Success 200 {string} string "updated"
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "server error"
// @Router /employees [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, uploader.MaxFileSize)
	if err := r.ParseMultipartForm(uploader.MaxFileSize); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid form data")
		return
	}

	// Save new photo if provided
	_, fh, _ := r.FormFile("photo")
	photoPath, err := uploader.SavePhoto(fh)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	photoURL := uploader.FullURL(r, photoPath)

	req := UpdateEmployeeRequest{
		FullName:    r.FormValue("full_name"),
		Email:       r.FormValue("email"),
		Birthday:    r.FormValue("birthday"),
		PhoneNumber: r.FormValue("phone_number"),
		Photo:       photoURL,
	}

	if v := r.FormValue("gender_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 32)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "gender_id must be a positive integer")
			return
		}
		req.GenderID = uint(id)
	}

	if req.PhoneNumber != "" {
		if err := validatePhoneNumber(req.PhoneNumber); err != nil {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	userID := uint(middleware.GetUserID(r))
	if err := h.service.Update(userID, req); err != nil {
		logger.Log.Error(err.Error())
		if err.Error() == "employee not found" {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteEmployee godoc
// @Summary Delete employee
// @Tags Employee
// @Security BearerAuth
// @Success 200 {string} string "deleted"
// @Failure 500 {string} string "server error"
// @Router /employees [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := uint(middleware.GetUserID(r))

	if err := h.service.Delete(userID); err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetActivities godoc
// @Summary Get employee activity contributions
// @Description Returns daily contribution counts (tasks created + comments), total contributions, and total active days for the authenticated employee.
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Success 200 {object} ActivitiesResponse
// @Failure 500 {string} string "server error"
// @Router /employees/me/activities [get]
func (h *Handler) GetActivities(w http.ResponseWriter, r *http.Request) {
	userID := uint(middleware.GetUserID(r))

	result, err := h.service.GetActivities(userID)
	if err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GetProfile godoc
// @Summary Get employee profile + activity dashboard
// @Description Returns profile info and activity dashboard for any employee by their ID (employee_id == user_id).
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Param id path int true "Employee ID (same as user_id)"
// @Success 200 {object} ProfileResponse
// @Failure 404 {string} string "not found"
// @Router /employees/{id}/profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := h.service.GetProfile(id)
	if err != nil {
		if err.Error() == "employee not found" {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	result.Profile.Photo = uploader.FullURL(r, result.Profile.Photo)
	response.JSON(w, http.StatusOK, result)
}

// GetAchievements godoc
// @Summary Get employee achievement progress
// @Description Returns the unlocked level for all 19 achievements for the given employee. Level 0 = locked, 1/2/3 = bronze/silver/gold, 4 = prestige.
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {array} AchievementResponse
// @Failure 400 {string} string "invalid id"
// @Failure 500 {string} string "server error"
// @Router /employees/{id}/achievements [get]
func (h *Handler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	result, err := h.service.GetAchievements(id)
	if err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// ExistsEmployee godoc
// @Summary Check if employee profile exists
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]bool
// @Router /employees/exists [get]
func (h *Handler) Exists(w http.ResponseWriter, r *http.Request) {
	userID := uint(middleware.GetUserID(r))

	res, err := h.service.Exists(userID)
	if err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]bool{"exists": res})
}

// --- helpers ---

func parseUintField(r *http.Request, field string) (uint, error) {
	v := r.FormValue(field)
	if v == "" {
		return 0, nil
	}
	id, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

func validateCreateRequest(req CreateEmployeeRequest) error {
	if req.FullName == "" {
		return errField("full_name is required")
	}
	if req.Email == "" {
		return errField("email is required")
	}
	if req.Birthday == "" {
		return errField("birthday is required")
	}
	if req.GenderID == 0 {
		return errField("gender_id is required")
	}
	if req.PhoneNumber != "" {
		if err := validatePhoneNumber(req.PhoneNumber); err != nil {
			return err
		}
	}
	return nil
}

func validatePhoneNumber(phone string) error {
	if len(phone) != 11 {
		return errField("phone_number must be exactly 11 digits")
	}
	for _, c := range phone {
		if c < '0' || c > '9' {
			return errField("phone_number must contain only digits")
		}
	}
	if phone[0] != '8' {
		return errField("phone_number must start with 8")
	}
	return nil
}

type fieldError string

func (e fieldError) Error() string { return string(e) }

func errField(msg string) error { return fieldError(msg) }

func parseID(r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
