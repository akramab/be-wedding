package user

import (
	"net/http"

	"be-wedding/internal/rest/response"
)

type RemindUserSendWeddingVideoResponse struct {
	Message string `json:"message"`
}

func (handler *userHandler) RemindUserSendWeddingVideo(w http.ResponseWriter, r *http.Request) {
	// ctx := r.Context()
	// userID := chi.URLParam(r, "id")

	resp := RemindUserSendWeddingVideoResponse{
		Message: "reminder and detailed info about the video will be sent to the provided WhatsApp number",
	}

	response.Respond(w, http.StatusCreated, resp)
}
