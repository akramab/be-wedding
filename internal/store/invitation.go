package store

import (
	"context"
	"database/sql"
	"time"
)

const (
	InvitationStatusAvailable = "AVAILABLE"
	InvitationStatusUsed      = "USED"

	InvitationTypeSingle = "SINGLE"
	InvitationTypeGroup  = "GROUP"

	InvitationSession1ID = "ccd52961-fa4e-43ba-a6df-a4c97849d899"
	InvitationSession2ID = "ccd52961-fa4e-43ba-a6df-a4c97849d898"
)

type InvitationData struct {
	ID       string
	Type     string
	Name     string
	Status   string
	Session  int64
	Schedule string

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type InvitationUserData struct {
	ID             string
	Name           string
	WhatsAppNumber string
	PeopleCount    int64
	Status         string
	QRImage        string
}

type InvitationCompleteData struct {
	Invitation InvitationData
	User       InvitationUserData
}

type Invitation interface {
	Insert(ctx context.Context, invitation *InvitationData) error
	FindOneByID(ctx context.Context, id string) (*InvitationData, error)
	FindOneCompleteDataByID(ctx context.Context, id string) (*InvitationCompleteData, error)
}
