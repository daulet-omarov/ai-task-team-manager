package chat

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

func boardIDFromURL(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "boardId"), 10, 64)
	return uint(id), err
}

func msgIDFromURL(r *http.Request) (uint, error) {
	id, err := strconv.ParseUint(chi.URLParam(r, "msgId"), 10, 64)
	return uint(id), err
}

// GetMessages godoc
// @Summary      Get board chat messages
// @Description  Returns paginated messages for a board (newest first). Members only.
// @Tags         Chat
// @Security     BearerAuth
// @Produce      json
// @Param        boardId  path      int  true  "Board ID"
// @Param        limit    query     int  false "Page size (default 50, max 100)"
// @Param        offset   query     int  false "Offset"
// @Success      200  {array}   MessageResponse
// @Router       /boards/{boardId}/chat [get]
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}
	userID := middleware.GetUserID(r)

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	msgs, err := h.service.GetMessages(boardID, userID, limit, offset)
	if err != nil {
		if err.Error() == "access denied: not a board member" {
			response.Error(w, http.StatusForbidden, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]MessageResponse, len(msgs))
	for i, m := range msgs {
		resp[i] = BuildMessageResponse(m, r)
	}
	response.JSON(w, http.StatusOK, resp)
}

// SendMessage godoc
// @Summary      Send a chat message
// @Description  Send a text / file / voice / photo / video message. Use multipart/form-data. Field "files" accepts multiple files.
// @Tags         Chat
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        boardId     path      int     true  "Board ID"
// @Param        text        formData  string  false "Message text"
// @Param        reply_to_id formData  int     false "ID of the message being replied to"
// @Param        files       formData  file    false "Attachments (repeat for multiple)"
// @Success      201  {object}  MessageResponse
// @Router       /boards/{boardId}/chat [post]
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}
	userID := middleware.GetUserID(r)

	ParseMultipartOrEmpty(r)

	var req SendMessageRequest
	req.Text = r.FormValue("text")
	if v := r.FormValue("reply_to_id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			req.ReplyToID = &uid
		}
	}

	msg, err := h.service.SendMessage(boardID, userID, r, req)
	if err != nil {
		switch err.Error() {
		case "access denied: not a board member":
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	built := BuildMessageResponse(*msg, r)
	h.hub.Broadcast(boardID, hub.Event{Type: "new_message", Data: built})
	response.JSON(w, http.StatusCreated, built)
}

// CreatePoll godoc
// @Summary      Create a poll message
// @Description  Creates a poll in the board chat. Members only.
// @Tags         Chat
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        boardId  path      int               true  "Board ID"
// @Param        request  body      CreatePollRequest true  "Poll data"
// @Success      201  {object}  MessageResponse
// @Router       /boards/{boardId}/chat/polls [post]
func (h *Handler) CreatePoll(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}
	userID := middleware.GetUserID(r)

	var req CreatePollRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	msg, err := h.service.CreatePoll(boardID, userID, req)
	if err != nil {
		switch err.Error() {
		case "access denied: not a board member":
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	built := BuildMessageResponse(*msg, r)
	h.hub.Broadcast(boardID, hub.Event{Type: "new_message", Data: built})
	response.JSON(w, http.StatusCreated, built)
}

// Vote godoc
// @Summary      Vote on a poll option (toggle)
// @Description  Adds a vote if not yet voted; removes it if already voted.
// @Tags         Chat
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        boardId  path      int         true  "Board ID"
// @Param        request  body      VoteRequest true  "Option to vote for"
// @Success      200  {object}  PollResponse
// @Router       /boards/{boardId}/chat/polls/vote [post]
func (h *Handler) Vote(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}
	userID := middleware.GetUserID(r)

	var req VoteRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	poll, err := h.service.Vote(boardID, userID, req)
	if err != nil {
		switch err.Error() {
		case "access denied: not a board member":
			response.Error(w, http.StatusForbidden, err.Error())
		case "poll option not found", "poll not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	built := BuildPollResponse(*poll)
	h.hub.Broadcast(boardID, hub.Event{Type: "update_poll", Data: built})
	response.JSON(w, http.StatusOK, built)
}

// Unvote godoc
// @Summary Remove vote from a poll option
// @Tags Chat
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param boardId path int true "Board ID"
// @Param request body VoteRequest true "Option to unvote"
// @Success 200 {object} PollResponse
// @Router /boards/{boardId}/chat/polls/unvote [post]
func (h *Handler) Unvote(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}
	userID := middleware.GetUserID(r)

	var req VoteRequest
	if err := request.DecodeAndValidate(r, &req); err != nil {
		response.Error(w, http.StatusBadRequest, err.Error())
		return
	}

	poll, err := h.service.Unvote(boardID, userID, req)
	if err != nil {
		switch err.Error() {
		case "access denied: not a board member":
			response.Error(w, http.StatusForbidden, err.Error())
		case "poll option not found", "poll not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "you have not voted for this option":
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	built := BuildPollResponse(*poll)
	h.hub.Broadcast(boardID, hub.Event{Type: "update_poll", Data: built})
	response.JSON(w, http.StatusOK, built)
}

// DeleteMessage godoc
// @Summary      Delete a chat message
// @Description  Deletes the message. Only the author can delete their own message.
// @Tags         Chat
// @Security     BearerAuth
// @Produce      json
// @Param        boardId  path  int  true  "Board ID"
// @Param        msgId    path  int  true  "Message ID"
// @Success      204
// @Router       /boards/{boardId}/chat/{msgId} [delete]
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}
	msgID, err := msgIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid message id")
		return
	}
	userID := middleware.GetUserID(r)

	if err := h.service.DeleteMessage(boardID, userID, msgID); err != nil {
		switch err.Error() {
		case "access denied: not a board member", "access denied: not the message author":
			response.Error(w, http.StatusForbidden, err.Error())
		case "message not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	h.hub.Broadcast(boardID, hub.Event{Type: "delete_message", Data: map[string]uint{"id": msgID}})
	w.WriteHeader(http.StatusNoContent)
}

// ServeWS godoc
// @Summary      WebSocket endpoint for real-time board events (chat + tasks + statuses)
// @Description  Connect via WS. Pass JWT as ?token=... query param.
// @Description  Receives all board events: new_message, delete_message, update_poll,
// @Description  task_created, task_updated, task_deleted,
// @Description  status_created, status_updated, status_deleted, statuses_reordered,
// @Description  status_default_changed, member_removed.
// @Tags         Chat
// @Param        boardId  path   int     true  "Board ID"
// @Param        token    query  string  true  "JWT token"
// @Router       /boards/{boardId}/ws [get]
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	boardID, err := boardIDFromURL(r)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid board id")
		return
	}

	token := r.URL.Query().Get("token")
	userID, err := pkgjwt.ParseToken(token)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "invalid token")
		return
	}

	if err := h.service.checkMember(boardID, userID); err != nil {
		response.Error(w, http.StatusForbidden, err.Error())
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	h.hub.Connect(conn, boardID)
}
