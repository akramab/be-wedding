package user

import (
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

type GetUserSpecificCommentResponse struct {
	CommentID string `json:"comment_id"`
	Comment   string `json:"comment"`
}

func (handler *userHandler) GetUserSpecificComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	userSpecificComment, err := handler.userStore.FindOneCommentByUserID(ctx, userID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("user comment not found"))
		return
	}

	resp := GetUserSpecificCommentResponse{
		CommentID: userSpecificComment.ID,
		Comment:   userSpecificComment.Comment,
	}

	response.Respond(w, http.StatusOK, resp)
}
