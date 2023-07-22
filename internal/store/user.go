package store

import (
	"context"
	"database/sql"
	"time"
)

const (
	UserStatusNewlyCreated  = "NEWLY_CREATED"
	UserStatusInfoCompleted = "INFO_COMPLETED"
	UserStatusRSVPProvided  = "RSVP_PROVIDED"
)

type UserData struct {
	ID             string
	InvitationID   string
	InvitationType string
	WhatsAppNumber string
	Name           string
	Status         string
	QRImageName    string

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type UserCommentData struct {
	ID       string
	UserID   string
	UserName string
	Comment  string
	Like     int64

	CreatedAt time.Time
	UpdatedAt sql.NullTime
}

type UserCommentLikeData struct {
	ID        string
	UserID    string
	CommentID string
	CreatedAt time.Time
}

type UserCommentLikeCountData struct {
	CommentID string
	LikeCount int64
}

type UserRSVPData struct {
	ID          string
	UserID      string
	PeopleCount int64
	IsAttending bool
	CreatedAt   time.Time
}

type User interface {
	Insert(ctx context.Context, user *UserData) error
	InsertWithID(ctx context.Context, user *UserData) error
	Update(ctx context.Context, user *UserData) error
	InsertComment(ctx context.Context, userComment *UserCommentData) error
	UpdateComment(ctx context.Context, userComment *UserCommentData) error
	FindAllComment(ctx context.Context, startDate string, endDate string) ([]*UserCommentData, error)
	FindAllCommentPagination(ctx context.Context, offset int, limit int, startDate string, endDate string) ([]*UserCommentData, error)
	LikeUnlikeComment(ctx context.Context, userID string, commentID string) (bool, error)
	FindLikedCommentOnlyByUserID(ctx context.Context, userID string) ([]*UserCommentLikeData, error)
	FindLikedCommentCount(ctx context.Context) ([]*UserCommentLikeCountData, error)
	FindOneCommentByUserID(ctx context.Context, userID string) (*UserCommentData, error)
	InsertUserRSVP(ctx context.Context, userRSVP *UserRSVPData) error
	UpdateRSVPByUserID(ctx context.Context, userRSVP *UserRSVPData) error
	FindAllWhatsAppNumber(ctx context.Context) ([]string, error)
	UpdateRSVPAttendanceByUserID(ctx context.Context, userRSVP *UserRSVPData) error
}
