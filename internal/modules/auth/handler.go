package auth

import (
	"net/http"

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

// Register godoc
// @Summary Register user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Register request"
// @Success 201 {string} string "created"
// @Failure 400 {string} string "bad request"
// @Failure 409 {string} string "email already exists"
// @Failure 500 {string} string "server error"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest

	err := request.DecodeAndValidate(r, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.Register(req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusConflict, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// Login godoc
// @Summary Login user
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login request"
// @Success 200 {object} map[string]string "token response"
// @Failure 400 {string} string "bad request"
// @Failure 401 {string} string "invalid credentials"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest

	err := request.DecodeAndValidate(r, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"token": token,
	})
}

// ForgotPassword godoc
// @Summary Forgot password
// @Description Check if email exists and initiate password reset
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body ForgotPasswordRequest true "Forgot password request"
// @Success 200 {string} string "ok"
// @Failure 400 {string} string "bad request"
// @Failure 404 {string} string "user not found"
// @Router /auth/forgot-password [post]
func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {

	var req ForgotPasswordRequest

	err := request.DecodeAndValidate(r, &req)
	if err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	err = h.service.ForgotPassword(req.Email)
	if err != nil {
		response.Error(w, http.StatusNotFound, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteAccount godoc
// @Summary Delete user account
// @Description Delete the authenticated user's account
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {string} string "account deleted"
// @Failure 401 {string} string "unauthorized"
// @Failure 500 {string} string "server error"
// @Router /auth/account [delete]
func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {

	userID := middleware.GetUserID(r)

	err := h.service.DeleteAccount(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
}
