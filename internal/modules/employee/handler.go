package employee

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/logger"
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

// CreateEmployee godoc
// @Summary Create employee
// @Description Create a new employee
// @Tags Employee
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateEmployeeRequest true "Create employee request"
// @Success 201 {string} string "created"
// @Failure 400 {string} string "bad request"
// @Failure 500 {string} string "server error"
// @Router /employees [post]
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateEmployeeRequest

	if err := request.DecodeAndValidate(r, &req); err != nil {
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
// @Summary Get employee by ID
// @Description Get a single employee by their ID
// @Tags Employee
// @Security BearerAuth
// @Produce json
// @Param id path int true "Employee ID"
// @Success 200 {object} EmployeeResponse
// @Failure 400 {string} string "invalid id"
// @Failure 404 {string} string "not found"
// @Router /employees/{id} [get]
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	emp, err := h.service.GetByID(id)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, emp)
}

// GetAllEmployees godoc
// @Summary Get all employees
// @Description Get list of all employees
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

	response.JSON(w, http.StatusOK, employees)
}

// UpdateEmployee godoc
// @Summary Update employee
// @Description Update an existing employee by ID
// @Tags Employee
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Employee ID"
// @Param request body UpdateEmployeeRequest true "Update employee request"
// @Success 200 {string} string "updated"
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "not found"
// @Failure 500 {string} string "server error"
// @Router /employees/{id} [put]
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	var req UpdateEmployeeRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.service.Update(id, req); err != nil {
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
// @Description Delete an employee by ID
// @Tags Employee
// @Security BearerAuth
// @Param id path int true "Employee ID"
// @Success 200 {string} string "deleted"
// @Failure 400 {string} string "invalid id"
// @Failure 500 {string} string "server error"
// @Router /employees/{id} [delete]
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.service.Delete(id); err != nil {
		logger.Log.Error(err.Error())
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// --- helper ---

func parseID(r *http.Request) (uint, error) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
