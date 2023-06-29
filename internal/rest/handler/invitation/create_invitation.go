package invitation

import (
	"encoding/json"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/store"

	"be-wedding/internal/rest/response"
)

type InvitationResponse struct {
	InvitationID string `json:"invitation_id"`
	Name         string `json:"name"`
	Session      int64  `json:"session"`
}

type CreateInvitationRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Session int64  `json:"session"`
}

func (handler *invitationHandler) CreateInvitation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := CreateInvitationRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	newInvitation := store.InvitationData{
		Type:    req.Type,
		Name:    req.Name,
		Session: req.Session,
	}

	if err := handler.invitationStore.Insert(ctx, &newInvitation); err != nil {
		log.Println("error insert new invitation data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	resp := InvitationResponse{
		InvitationID: newInvitation.ID,
		Name:         newInvitation.Name,
		Session:      newInvitation.Session,
	}

	response.Respond(w, http.StatusCreated, resp)
}
