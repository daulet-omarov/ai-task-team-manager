package invite

import (
	"net/http"
	"strconv"

	"github.com/daulet-omarov/ai-task-team-manager/internal/hub"
	"github.com/daulet-omarov/ai-task-team-manager/internal/middleware"
	"github.com/daulet-omarov/ai-task-team-manager/internal/request"
	"github.com/daulet-omarov/ai-task-team-manager/internal/response"
	pkgjwt "github.com/daulet-omarov/ai-task-team-manager/pkg/jwt"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	service *Service
	hub     *hub.Hub
}

func NewHandler(service *Service, h *hub.Hub) *Handler {
	return &Handler{service: service, hub: h}
}

// Invite godoc
// @Summary Send invitation
// @Description Board owner invites a user to the board
// @Tags Invite
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param boardId path int true "Board ID"
// @Param request body CreateInviteRequest true "Invite request"
// @Success 201 {string} string "invitation sent"
// @Failure 400 {string} string "bad request"
// @Failure 403 {string} string "only the board owner can send invitations"
// @Failure 409 {string} string "already a member / invitation already pending"
// @Router /boards/{boardId}/invite [post]
func (h *Handler) Invite(w http.ResponseWriter, r *http.Request) {
	boardID, err := strconv.ParseUint(chi.URLParam(r, "boardId"), 10, 32)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	var req CreateInviteRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := middleware.GetUserID(r)

	inv, err := h.service.Invite(uint(boardID), userID, req)
	if err != nil {
		switch err.Error() {
		case "only the board owner can send invitations":
			response.Error(w, http.StatusForbidden, err.Error())
		case "user is already a board member", "invitation already pending":
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	h.hub.BroadcastToUser(req.UserID, hub.Event{Type: "invite_received", Data: inv})

	w.WriteHeader(http.StatusCreated)
}

// GetInvites godoc
// @Summary List my invitations
// @Description Returns all pending invitations for the authenticated user
// @Tags Invite
// @Security BearerAuth
// @Produce json
// @Success 200 {array} InviteResponse
// @Router /invites [get]
func (h *Handler) GetInvites(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)

	invites, err := h.service.GetUserInvites(userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, invites)
}

// Accept godoc
// @Summary Accept invitation
// @Description Accept a pending board invitation
// @Tags Invite
// @Security BearerAuth
// @Produce json
// @Param inviteId path int true "Invite ID"
// @Success 200 {string} string "accepted"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "invitation not found"
// @Router /invites/{inviteId}/accept [post]
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	inviteID, err := parseInviteID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid invite id")
		return
	}

	userID := middleware.GetUserID(r)

	boardID, err := h.service.Accept(inviteID, userID)
	if err != nil {
		switch err.Error() {
		case "access denied":
			response.Error(w, http.StatusForbidden, err.Error())
		case "invitation not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.hub.Broadcast(boardID, hub.Event{Type: "member_added", Data: map[string]int64{"user_id": userID}})

	w.WriteHeader(http.StatusOK)
}

// Reject godoc
// @Summary Reject invitation
// @Description Reject a pending board invitation
// @Tags Invite
// @Security BearerAuth
// @Produce json
// @Param inviteId path int true "Invite ID"
// @Success 200 {string} string "rejected"
// @Failure 403 {string} string "access denied"
// @Failure 404 {string} string "invitation not found"
// @Router /invites/{inviteId}/reject [post]
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	inviteID, err := parseInviteID(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid invite id")
		return
	}

	userID := middleware.GetUserID(r)

	if err := h.service.Reject(inviteID, userID); err != nil {
		switch err.Error() {
		case "access denied":
			response.Error(w, http.StatusForbidden, err.Error())
		case "invitation not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ConnectUserWS godoc
// @Summary WebSocket endpoint for real-time user events (invitations)
// @Description Connect via WS. Pass JWT as ?token=... query param.
// @Description Receives events: invite_received.
// @Tags Invite
// @Param token query string true "JWT token"
// @Router /users/ws [get]
func (h *Handler) ConnectUserWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	userID, err := pkgjwt.ParseToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid token")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.ConnectUser(conn, userID)
}

func parseInviteID(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "inviteId"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
