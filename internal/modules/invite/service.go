package invite

import (
	"errors"

	"github.com/daulet-omarov/ai-task-team-manager/internal/models"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Invite sends a board invitation. Only the board owner can invite users.
// Returns the created invitation so the caller can broadcast it to the invitee.
func (s *Service) Invite(boardID uint, inviterID int64, req CreateInviteRequest) (*InviteResponse, error) {
	isOwner, err := s.repo.IsOwner(boardID, inviterID)
	if err != nil {
		return nil, err
	}
	if !isOwner {
		return nil, errors.New("only the board owner can send invitations")
	}

	isMember, err := s.repo.IsMember(boardID, req.UserID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("user is already a board member")
	}

	exists, err := s.repo.ExistsPending(boardID, req.UserID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("invitation already pending")
	}

	inv := &models.Invite{
		BoardID:   boardID,
		InviterID: inviterID,
		InviteeID: req.UserID,
		Status:    models.InviteStatusPending,
	}

	if err := s.repo.Create(inv); err != nil {
		return nil, err
	}

	// Reload to get Board name populated.
	full, err := s.repo.GetByID(inv.ID)
	if err != nil {
		return nil, err
	}
	return toInviteResponse(full), nil
}

// GetUserInvites returns all pending invitations for the given user.
func (s *Service) GetUserInvites(userID int64) ([]*InviteResponse, error) {
	invites, err := s.repo.GetPendingByInvitee(userID)
	if err != nil {
		return nil, err
	}

	result := make([]*InviteResponse, 0, len(invites))
	for _, inv := range invites {
		result = append(result, toInviteResponse(inv))
	}

	return result, nil
}

func toInviteResponse(inv *models.Invite) *InviteResponse {
	return &InviteResponse{
		ID:        inv.ID,
		BoardID:   inv.BoardID,
		BoardName: inv.Board.Name,
		InviterID: inv.InviterID,
		InviteeID: inv.InviteeID,
		Status:    inv.Status,
		CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// Accept accepts a pending invitation and adds the user to the board.
// Returns the boardID so the caller can broadcast member_added to that board room.
func (s *Service) Accept(inviteID uint, userID int64) (uint, error) {
	inv, err := s.repo.GetByID(inviteID)
	if err != nil {
		return 0, errors.New("invitation not found")
	}
	if inv.InviteeID != userID {
		return 0, errors.New("access denied")
	}
	if inv.Status != models.InviteStatusPending {
		return 0, errors.New("invitation is no longer pending")
	}

	if err := s.repo.UpdateStatus(inviteID, models.InviteStatusAccepted); err != nil {
		return 0, err
	}

	return inv.BoardID, s.repo.AddMember(inv.BoardID, userID)
}

// Reject declines a pending invitation.
func (s *Service) Reject(inviteID uint, userID int64) error {
	inv, err := s.repo.GetByID(inviteID)
	if err != nil {
		return errors.New("invitation not found")
	}
	if inv.InviteeID != userID {
		return errors.New("access denied")
	}
	if inv.Status != models.InviteStatusPending {
		return errors.New("invitation is no longer pending")
	}

	return s.repo.UpdateStatus(inviteID, models.InviteStatusRejected)
}
