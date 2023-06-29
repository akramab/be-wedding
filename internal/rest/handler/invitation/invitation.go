package invitation

import (
	"database/sql"
	"net/http"

	"be-wedding/internal/config"
	"be-wedding/internal/store"
)

type InvitationHandler interface {
	CreateInvitation(w http.ResponseWriter, r *http.Request)
	GetInvitationCompleteData(w http.ResponseWriter, r *http.Request)
}

type invitationHandler struct {
	apiCfg          config.API
	db              *sql.DB
	invitationStore store.Invitation
}

func NewInvitationHandler(apiCfg config.API, db *sql.DB, invitationStore store.Invitation) InvitationHandler {
	return &invitationHandler{
		apiCfg:          apiCfg,
		db:              db,
		invitationStore: invitationStore,
	}
}
