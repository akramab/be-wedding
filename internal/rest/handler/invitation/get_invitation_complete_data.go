package invitation

import (
	"fmt"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

type InvidationData struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Status   string `json:"status"`
	Schedule string `json:"schedule"`
}

type UserData struct {
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	WhatsAppNumber string `json:"wa_number,omitempty"`
	Status         string `json:"status,omitempty"`
	QRImageLink    string `json:"qr_image_link,omitempty"`
	PeopleCount    int64  `json:"people_count,omitempty"`
}

type GetInvitationCompleteDataResponse struct {
	Invitation InvidationData `json:"invitation"`
	User       UserData       `json:"user,omitempty"`
}

func (handler *invitationHandler) GetInvitationCompleteData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	invitationID := chi.URLParam(r, "id")

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByID(ctx, invitationID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("Invitation id not found"))
		return
	}

	resp := GetInvitationCompleteDataResponse{
		Invitation: InvidationData{
			ID:       invitationCompleteData.Invitation.ID,
			Name:     invitationCompleteData.Invitation.Name,
			Type:     invitationCompleteData.Invitation.Type,
			Status:   invitationCompleteData.Invitation.Status,
			Schedule: invitationCompleteData.Invitation.Schedule,
		},
		User: UserData{
			ID:             invitationCompleteData.User.ID,
			Name:           invitationCompleteData.User.Name,
			WhatsAppNumber: invitationCompleteData.User.WhatsAppNumber,
			Status:         invitationCompleteData.User.Status,
			QRImageLink:    fmt.Sprintf("http://localhost/static/%s", invitationCompleteData.User.QRImage),
			PeopleCount:    invitationCompleteData.User.PeopleCount,
		},
	}

	response.Respond(w, http.StatusOK, resp)
}
