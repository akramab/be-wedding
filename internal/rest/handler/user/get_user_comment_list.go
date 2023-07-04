package user

import (
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"be-wedding/internal/rest/response"

	apierror "be-wedding/internal/rest/error"

	"github.com/go-chi/chi/v5"
)

type GetUserCommentListRequest struct {
	startDateStr string
	startDate    time.Time
	endDateStr   string
	endDate      time.Time
	pageStr      string
	limitStr     string
	page         int
	limit        int
}

func (r *GetUserCommentListRequest) validate() *apierror.FieldError {
	var err error
	fieldErr := apierror.NewFieldError()

	r.pageStr = strings.TrimSpace(r.pageStr)
	r.limitStr = strings.TrimSpace(r.limitStr)

	if r.pageStr == "" {
		r.pageStr = "1"
	}
	if r.limitStr == "" {
		r.limitStr = "50"
	}

	r.page, err = strconv.Atoi(r.pageStr)
	if err != nil || r.page < 0 {
		fieldErr = fieldErr.WithField("page", "page must be a positive integer")
	}

	r.limit, err = strconv.Atoi(r.limitStr)
	if err != nil || r.limit < 0 {
		fieldErr = fieldErr.WithField("limit", "limit must be a positive integer")
	}

	if r.startDateStr != "" {
		r.startDate, err = time.Parse(time.DateOnly, r.startDateStr)
		if err != nil {
			fieldErr = fieldErr.WithField("start_date", "start_date must be in the format of YYYY-MM-DD!")
		}
	}

	if r.endDateStr != "" {
		r.endDate, err = time.Parse(time.DateOnly, r.endDateStr)
		if err != nil {
			fieldErr = fieldErr.WithField("end_date", "end_date must be in the format of YYYY-MM-DD!")
		}
		dayInt, _ := strconv.Atoi(r.endDateStr[8:10])
		dayInt++
		if dayInt < 10 {
			r.endDateStr = r.endDateStr[0:8] + "0" + strconv.Itoa(dayInt)
		} else {
			r.endDateStr = r.endDateStr[0:8] + strconv.Itoa(dayInt)
		}
	}

	if len(fieldErr.Fields) != 0 {
		return &fieldErr
	}

	return nil
}

type GetUserCommentListResponse struct {
	TotalItems int               `json:"total_items"`
	TotalPages int               `json:"total_pages"`
	Items      []UserCommentItem `json:"items"`
}

type UserCommentItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Comment   string    `json:"comment"`
	LikeCount int64     `json:"like_count"`
	IsLiked   bool      `json:"is_liked"`
	CreatedAt time.Time `json:"created_at"`
}

func (handler *userHandler) GetUserCommentList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	req := GetUserCommentListRequest{
		pageStr:      r.URL.Query().Get("page"),
		limitStr:     r.URL.Query().Get("limit"),
		startDateStr: r.URL.Query().Get("start_date"),
		endDateStr:   r.URL.Query().Get("end_date"),
	}

	fieldErr := req.validate()
	if fieldErr != nil {
		response.FieldError(w, *fieldErr)
		return
	}

	userCommentListAll, err := handler.userStore.FindAllComment(ctx, req.startDateStr, req.endDateStr)
	if err != nil {
		log.Println(err)

		response.Error(w, apierror.InternalServerError())
		return
	}
	totalItems := len(userCommentListAll)

	offset := req.limit * (req.page - 1)
	userCommentList, err := handler.userStore.FindAllCommentPagination(ctx, offset, req.limit, req.startDateStr, req.endDateStr)
	if err != nil {
		log.Println(err)

		response.Error(w, apierror.InternalServerError())
		return
	}
	itemsCount := len(userCommentList)

	userCommentLikeList, err := handler.userStore.FindLikedCommentOnlyByUserID(ctx, userID)
	if err != nil {
		log.Println(err)

		response.Error(w, apierror.InternalServerError())
		return
	}

	resp := GetUserCommentListResponse{}
	items := make([]UserCommentItem, itemsCount)
	for idx, item := range userCommentList {
		var isLiked bool
		for _, userCommentLike := range userCommentLikeList {
			if item.ID == userCommentLike.CommentID {
				isLiked = true
				break
			}
		}

		items[idx] = UserCommentItem{
			ID:        item.ID,
			UserID:    item.UserID,
			UserName:  item.UserName,
			Comment:   item.Comment,
			LikeCount: item.Like,
			IsLiked:   isLiked,
			CreatedAt: item.CreatedAt,
		}
	}
	resp.TotalItems = totalItems
	resp.TotalPages = int(math.Ceil(float64(totalItems) / float64(req.limit)))
	resp.Items = items

	response.Respond(w, http.StatusOK, resp)
}
