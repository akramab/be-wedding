package user

import (
	"be-wedding/internal/rest/response"
	"net/http"
	"strconv"
	"time"
)

const (
	GetCurrentVideoList = "CURRENT_VIDE_LIST"
	GetCurrentIndex     = "CURRENT_INDEX"

	DefaultCacheTime = time.Duration(10) * time.Minute
)

type GetCurrentVideoResponse struct {
	VideoList    []string `json:"video_list"`
	CurrentIndex int      `json:"current_index"`
}

func (handler *userHandler) GetCurrentVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	videoUrlList := []string{}
	var currentIdx int
	videoResultString, _ := handler.redisCache.Get(ctx, GetCurrentVideoList).Result()
	currentIndexResultString, _ := handler.redisCache.Get(ctx, GetCurrentIndex).Result()

	if videoResultString == "" {
		videoUrlList = []string{
			"https://api.kramili.site/static/3.mp4",
			"https://api.kramili.site/static/5.mp4",
		}
	}

	if currentIndexResultString == "" {
		currentIdx = 0
	} else {
		currentIdx, _ = strconv.Atoi(currentIndexResultString)
		currentIdx++
	}
	if currentIdx >= len(videoUrlList) {
		currentIdx = 0
	}
	handler.redisCache.Set(ctx, GetCurrentIndex, currentIdx, DefaultCacheTime).Result()

	resp := GetCurrentVideoResponse{
		VideoList:    videoUrlList,
		CurrentIndex: currentIdx,
	}

	response.Respond(w, http.StatusCreated, resp)
}
