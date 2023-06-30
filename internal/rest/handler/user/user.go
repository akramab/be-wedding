package user

import (
	"database/sql"
	"net/http"

	"be-wedding/internal/config"
	"be-wedding/internal/store"
	"be-wedding/pkg/whatsapp"
)

type UserHandler interface {
	CreateUser(w http.ResponseWriter, r *http.Request)
	UpdateUser(w http.ResponseWriter, r *http.Request)
	CreateUserComment(w http.ResponseWriter, r *http.Request)
	UpdateUserComment(w http.ResponseWriter, r *http.Request)
	GetUserCommentList(w http.ResponseWriter, r *http.Request)
	LikeUnlikeComment(w http.ResponseWriter, r *http.Request)
	GetUserSpecificComment(w http.ResponseWriter, r *http.Request)
	CreateUserRSVP(w http.ResponseWriter, r *http.Request)
	RemindUserWeddingDate(w http.ResponseWriter, r *http.Request)
	RemindUserSendWeddingVideo(w http.ResponseWriter, r *http.Request)
}

type userHandler struct {
	apiCfg          config.API
	db              *sql.DB
	userStore       store.User
	invitationStore store.Invitation
	waClient        whatsapp.Client
}

func NewUserHandler(apiCfg config.API, db *sql.DB, userStore store.User, invitationStore store.Invitation, waClient whatsapp.Client) UserHandler {
	return &userHandler{
		apiCfg:          apiCfg,
		db:              db,
		userStore:       userStore,
		invitationStore: invitationStore,
		waClient: waClient,
	}
}
