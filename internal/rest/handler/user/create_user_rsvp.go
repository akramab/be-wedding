package user

import (
	"encoding/json"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/store"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

type CreateUserRSVPRequest struct {
	PeopleCount int64 `json:"people_count"`
}

type CreateUserRSVPResponse struct {
	Message     string `json:"message"`
	UserRSVPID  string `json:"user_rsvp_id"`
	PeopleCount int64  `json:"people_count"`
}

func (handler *userHandler) CreateUserRSVP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	req := CreateUserRSVPRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	userRSVP := &store.UserRSVPData{
		UserID:      userID,
		PeopleCount: req.PeopleCount,
	}

	if err := handler.userStore.InsertUserRSVP(ctx, userRSVP); err != nil {
		log.Println("error insert new user rsvp data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	resp := CreateUserRSVPResponse{
		Message:     "success",
		UserRSVPID:  userRSVP.ID,
		PeopleCount: userRSVP.PeopleCount,
	}

	response.Respond(w, http.StatusCreated, resp)
}
