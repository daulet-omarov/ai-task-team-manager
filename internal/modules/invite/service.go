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
func (s *Service) Invite(boardID uint, inviterID int64, req CreateInviteRequest) error {
	isOwner, err := s.repo.IsOwner(boardID, inviterID)
	if err != nil {
		return err
	}
	if !isOwner {
		return errors.New("only the board owner can send invitations")
	}

	isMember, err := s.repo.IsMember(boardID, req.UserID)
	if err != nil {
		return err
	}
	if isMember {
		return errors.New("user is already a board member")
	}

	exists, err := s.repo.ExistsPending(boardID, req.UserID)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("invitation already pending")
	}

	inv := &models.Invite{
		BoardID:   boardID,
		InviterID: inviterID,
		InviteeID: req.UserID,
		Status:    models.InviteStatusPending,
	}

	return s.repo.Create(inv)
}

// GetUserInvites returns all pending invitations for the given user.
func (s *Service) GetUserInvites(userID int64) ([]*InviteResponse, error) {
	invites, err := s.repo.GetPendingByInvitee(userID)
	if err != nil {
		return nil, err
	}

	result := make([]*InviteResponse, 0, len(invites))
	for _, inv := range invites {
		result = append(result, &InviteResponse{
			ID:        inv.ID,
			BoardID:   inv.BoardID,
			BoardName: inv.Board.Name,
			InviterID: inv.InviterID,
			InviteeID: inv.InviteeID,
			Status:    inv.Status,
			CreatedAt: inv.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return result, nil
}

// Accept accepts a pending invitation and adds the user to the board.
func (s *Service) Accept(inviteID uint, userID int64) error {
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

	if err := s.repo.UpdateStatus(inviteID, models.InviteStatusAccepted); err != nil {
		return err
	}

	return s.repo.AddMember(inv.BoardID, userID)
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
