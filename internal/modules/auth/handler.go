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

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {

	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = h.service.Register(req.Email, req.Password)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {

	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	err = h.service.Login(req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", 401)
		return
	}

	w.WriteHeader(http.StatusOK)
}

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

func (h *Handler) DeleteAccount(w http.ResponseWriter, r *http.Request) {

	userID := int64(1) // позже из JWT

	err := h.service.DeleteAccount(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}
