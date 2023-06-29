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

type CreateUserCommentRequest struct {
	Comment string `json:"comment"`
}

type CreateUserCommentResponse struct {
	Message   string `json:"message"`
	CommentID string `json:"comment_id"`
	Comment   string `json:"comment"`
}

func (handler *userHandler) CreateUserComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	req := CreateUserCommentRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	newUserCommentData := &store.UserCommentData{
		UserID:  userID,
		Comment: req.Comment,
	}

	if err := handler.userStore.InsertComment(ctx, newUserCommentData); err != nil {
		log.Println("error insert new user comment data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	resp := CreateUserCommentResponse{
		Message:   "success",
		CommentID: newUserCommentData.ID,
		Comment:   newUserCommentData.Comment,
	}

	response.Respond(w, http.StatusCreated, resp)
}
