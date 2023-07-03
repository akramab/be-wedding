package pgsql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"be-wedding/internal/store"

	"github.com/dchest/uniuri"
)

type Invitation struct {
	db *sql.DB
}

func NewInvitation(db *sql.DB) *Invitation {
	return &Invitation{db: db}
}

const invitationInsert = `INSERT INTO
invitations(
	id, session_id, type, name, status, created_at
) values(
	$1, $2, $3, $4, $5, $6
)
`

func (s *Invitation) Insert(ctx context.Context, invitation *store.InvitationData) error {
	insertStmt, err := s.db.PrepareContext(ctx, invitationInsert)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	var sessionID string
	if invitation.Session == 1 {
		sessionID = store.InvitationSession1ID
	} else {
		sessionID = store.InvitationSession2ID

	}
	invitationID := uniuri.NewLen(6)
	createdAt := time.Now().UTC()

	invitationStatus := store.InvitationStatusAvailable
	_, err = tx.StmtContext(ctx, insertStmt).ExecContext(ctx,
		invitationID, sessionID, invitation.Type, invitation.Name, invitationStatus, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}
	invitation.ID = invitationID
	invitation.Status = invitationStatus
	invitation.CreatedAt = createdAt

	return nil
}

const invitationFindOneByIDQuery = `SELECT id, type, name, status
		FROM invitations WHERE id = $1
	`

func (s *Invitation) FindOneByID(ctx context.Context, id string) (*store.InvitationData, error) {
	invitation := &store.InvitationData{}

	row := s.db.QueryRowContext(ctx, invitationFindOneByIDQuery, id)

	err := row.Scan(
		&invitation.ID, &invitation.Type, &invitation.Name, &invitation.Status,
	)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

const invitationFindOneCompleteDataByIDQuery = `SELECT i.id, i.type, i.name, i.status, invs.schedule, COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.wa_number, ''), COALESCE(u.status, ''), COALESCE(u.qr_image, ''), COALESCE(ursvp.people_count, 0), COALESCE(u.is_date_reminder_sent, FALSE), COALESCE(u.is_video_reminder_sent, FALSE)
		FROM invitations i
		LEFT JOIN invitation_sessions invs
		ON i.session_id = invs.id
		LEFT JOIN users u
		ON i.id = u.invitation_id
		LEFT JOIN user_rsvps ursvp
		ON u.id = ursvp.user_id
		WHERE i.id = $1
		LIMIT 1
	`

func (s *Invitation) FindOneCompleteDataByID(ctx context.Context, id string) (*store.InvitationCompleteData, error) {
	invitation := &store.InvitationCompleteData{}

	row := s.db.QueryRowContext(ctx, invitationFindOneCompleteDataByIDQuery, id)

	err := row.Scan(
		&invitation.Invitation.ID, &invitation.Invitation.Type,
		&invitation.Invitation.Name, &invitation.Invitation.Status, &invitation.Invitation.Schedule,
		&invitation.User.ID, &invitation.User.Name, &invitation.User.WhatsAppNumber, &invitation.User.Status,
		&invitation.User.QRImage, &invitation.User.PeopleCount, &invitation.User.IsDateReminderSent, &invitation.User.IsVideoReminderSent,
	)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

const invitationFindOneCompleteDataByUserIDQuery = `SELECT i.id, i.type, i.name, i.status, invs.schedule, COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.wa_number, ''), COALESCE(u.status, ''), COALESCE(u.qr_image, ''), COALESCE(ursvp.people_count, 0), COALESCE(u.is_date_reminder_sent, FALSE), COALESCE(u.is_video_reminder_sent, FALSE)
		FROM invitations i
		LEFT JOIN invitation_sessions invs
		ON i.session_id = invs.id
		LEFT JOIN users u
		ON i.id = u.invitation_id
		LEFT JOIN user_rsvps ursvp
		ON u.id = ursvp.user_id
		WHERE u.id = $1
		LIMIT 1
	`

func (s *Invitation) FindOneCompleteDataByUserID(ctx context.Context, id string) (*store.InvitationCompleteData, error) {
	invitation := &store.InvitationCompleteData{}

	row := s.db.QueryRowContext(ctx, invitationFindOneCompleteDataByUserIDQuery, id)

	err := row.Scan(
		&invitation.Invitation.ID, &invitation.Invitation.Type,
		&invitation.Invitation.Name, &invitation.Invitation.Status, &invitation.Invitation.Schedule,
		&invitation.User.ID, &invitation.User.Name, &invitation.User.WhatsAppNumber, &invitation.User.Status,
		&invitation.User.QRImage, &invitation.User.PeopleCount, &invitation.User.IsDateReminderSent, &invitation.User.IsVideoReminderSent,
	)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

const invitationFindOneCompleteDataByWANumberQuery = `SELECT i.id, i.type, i.name, i.status, invs.schedule, COALESCE(u.id, ''), COALESCE(u.name, ''), COALESCE(u.wa_number, ''), COALESCE(u.status, ''), COALESCE(u.qr_image, ''), COALESCE(ursvp.people_count, 0), COALESCE(u.is_date_reminder_sent, FALSE), COALESCE(u.is_video_reminder_sent, FALSE)
		FROM invitations i
		LEFT JOIN invitation_sessions invs
		ON i.session_id = invs.id
		LEFT JOIN users u
		ON i.id = u.invitation_id
		LEFT JOIN user_rsvps ursvp
		ON u.id = ursvp.user_id
		WHERE u.wa_number = $1
		LIMIT 1
	`

func (s *Invitation) FindOneCompleteDataByWANumber(ctx context.Context, waNumber string) (*store.InvitationCompleteData, error) {
	invitation := &store.InvitationCompleteData{}

	row := s.db.QueryRowContext(ctx, invitationFindOneCompleteDataByWANumberQuery, waNumber)

	err := row.Scan(
		&invitation.Invitation.ID, &invitation.Invitation.Type,
		&invitation.Invitation.Name, &invitation.Invitation.Status, &invitation.Invitation.Schedule,
		&invitation.User.ID, &invitation.User.Name, &invitation.User.WhatsAppNumber, &invitation.User.Status,
		&invitation.User.QRImage, &invitation.User.PeopleCount, &invitation.User.IsDateReminderSent, &invitation.User.IsVideoReminderSent,
	)
	if err != nil {
		return nil, err
	}

	return invitation, nil
}

const userUpdateReminderDateQuery = `UPDATE users
	SET is_date_reminder_sent = $2, updated_at = $3
	WHERE id = $1
`

func (i *Invitation) UpdateDateReminder(ctx context.Context, invitationCompleteData *store.InvitationCompleteData) error {
	updateStmt, err := i.db.PrepareContext(ctx, userUpdateReminderDateQuery)
	if err != nil {
		return err
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updatedAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		invitationCompleteData.User.ID, !invitationCompleteData.User.IsDateReminderSent, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	invitationCompleteData.User.IsDateReminderSent = !invitationCompleteData.User.IsDateReminderSent

	return nil

}

const userUpdateReminderVideoQuery = `UPDATE users
	SET is_video_reminder_sent = $2, updated_at = $3
	WHERE id = $1
`

func (i *Invitation) UpdateVideoReminder(ctx context.Context, invitationCompleteData *store.InvitationCompleteData) error {
	updateStmt, err := i.db.PrepareContext(ctx, userUpdateReminderVideoQuery)
	if err != nil {
		return err
	}
	tx, err := i.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback()

	updatedAt := time.Now().UTC()
	_, err = tx.StmtContext(ctx, updateStmt).ExecContext(ctx,
		invitationCompleteData.User.ID, !invitationCompleteData.User.IsVideoReminderSent, updatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	invitationCompleteData.User.IsVideoReminderSent = !invitationCompleteData.User.IsVideoReminderSent

	return nil

}
