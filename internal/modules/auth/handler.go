package auth

import (
	"encoding/json"
	"net/http"
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		http.Error(w, "email and password required", http.StatusBadRequest)
		return
	}

	err = h.service.Register(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	token, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}

	resp := map[string]string{
		"token": token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = h.service.ForgotPassword(req.Email)
	if err != nil {
		http.Error(w, err.Error(), 404)
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

	userID := GetUserID(r)

	err := h.service.DeleteAccount(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}
