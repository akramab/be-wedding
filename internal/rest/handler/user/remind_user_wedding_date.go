package user

import (
	"net/http"

	"be-wedding/internal/rest/response"
)

type RemindUserWeddingDateResponse struct {
	Message string `json:"message"`
}

func (handler *userHandler) RemindUserWeddingDate(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// userID := chi.URLParam(r, "id")

	resp := RemindUserWeddingDateResponse{
		Message: "reminder will be sent to the provided WhatsApp number",
	}

	response.Respond(w, http.StatusCreated, resp)
}
