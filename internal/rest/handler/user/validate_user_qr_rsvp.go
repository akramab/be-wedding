package user

import (
	"be-wedding/internal/rest/response"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ValidateUserQRRsvpResponse struct {
	Name        string `json:"name"`
	PeopleCount int    `json:"people_count"`
	VIP         bool   `json:"vip"`
}

func (handler *userHandler) ValidateUserQRRsvp(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	invitationCode := chi.URLParam(r, "invitation_code")

	if invitationCode == "0c467423-e324-4786-a9d9-0c77eb267407" {
		resp := ValidateUserQRRsvpResponse{
			Name:        "Bapak Ari",
			PeopleCount: 2,
			VIP:         true,
		}

		response.Respond(w, http.StatusOK, resp)
		return
	}

	if invitationCode == "69fee15d-2c45-48ea-982e-4ce6327298fc" {
		resp := ValidateUserQRRsvpResponse{
			Name:        "Bapak Lukman",
			PeopleCount: 10,
			VIP:         false,
		}

		response.Respond(w, http.StatusOK, resp)
		return
	}

	resp := ValidateUserQRRsvpResponse{
		Name:        "Tamu Tidak Diundang",
		PeopleCount: -99999,
		VIP:         false,
	}

	response.Respond(w, http.StatusOK, resp)
	return
}
