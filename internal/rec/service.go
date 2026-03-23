package rec

import (
	"context"
	"database/sql"
	"errors"

	"github.com/akbarsenawijaya/solar-forecast/internal/notification"
	"github.com/akbarsenawijaya/solar-forecast/internal/user"
	"github.com/google/uuid"
)

type service struct {
	repo    Repository
	userSvc user.Service
	notif   notification.Service
}

func NewService(repo Repository, userSvc user.Service, notif notification.Service) Service {
	return &service{repo: repo, userSvc: userSvc, notif: notif}
}

func (s *service) UpdateAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID, kwh float64) error {
	newTotalKwh, err := s.repo.UpsertAccumulator(ctx, userID, profileID, kwh)
	if err != nil {
		return err
	}

	acc, err := s.GetAccumulator(ctx, userID, profileID)
	if err != nil {
		return err
	}

	// Calculate total MWh. 1 REC = 1 MWh (1000 kWh). Check if milestone reached.
	newTotalMwh := newTotalKwh / 1000

	if newTotalMwh >= 1.0 && !acc.MilestoneReached {
		// Set milestone to true
		_ = s.repo.SetMilestoneReached(ctx, userID, profileID)

		// Fetch user to get email and name
		if u, err := s.userSvc.GetUserByID(userID); err == nil {
			// Trigger milestone notification email
			_ = s.notif.SendRECMilestoneEmail(u.Email, u.Name, newTotalMwh)
		}
	}

	return nil
}

func (s *service) GetAccumulator(ctx context.Context, userID uuid.UUID, profileID *uuid.UUID) (*MwhAccumulator, error) {
	acc, err := s.repo.GetAccumulator(ctx, userID, profileID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &MwhAccumulator{
				UserID:         userID,
				SolarProfileID: profileID,
				CumulativeKwh:  0,
			}, nil
		}
		return nil, err
	}
	return acc, nil
}

func (s *service) GetTotalMwhForUser(ctx context.Context, userID uuid.UUID) (float64, error) {
	kwh, err := s.repo.GetTotalKwhForUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	return kwh / 1000, nil
}
