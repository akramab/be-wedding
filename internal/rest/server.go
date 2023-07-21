package rest

import (
	"be-wedding/internal/config"
	invitationhandler "be-wedding/internal/rest/handler/invitation"
	userhandler "be-wedding/internal/rest/handler/user"
	"be-wedding/internal/rest/middleware"
	storepgsql "be-wedding/internal/store/pgsql"
	"be-wedding/pkg/redis"
	"be-wedding/pkg/token"
	"be-wedding/pkg/whatsapp"
	"database/sql"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func New(
	cfg *config.Config,
	zlogger zerolog.Logger,
	sqlDB *sql.DB,
	redisCache redis.Client,
) http.Handler {
	r := chi.NewRouter()

	r.Use(
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"https://*", "http://*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		}),
		middleware.HTTPTracer,
		middleware.RequestID(zlogger),
		middleware.HTTPLogger,
	)

	invitationStore := storepgsql.NewInvitation(sqlDB)
	userStore := storepgsql.NewUser(sqlDB)

	whatsAppClient, whatsAppClientErr := whatsapp.NewWhatsMeowClient(cfg.WhatsApp, userStore, invitationStore, redisCache)
	if whatsAppClientErr != nil {
		zlogger.Error().Err(whatsAppClientErr).Msgf("rest: main failed to construct WhatsApp client: %s", whatsAppClientErr)
		return nil
	}

	invitationHandler := invitationhandler.NewInvitationHandler(cfg.API, sqlDB, invitationStore)
	userHandler := userhandler.NewUserHandler(cfg.API, cfg.WhatsApp, sqlDB, userStore, invitationStore, whatsAppClient, redisCache)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Afra & Akram Wedding Backend APIs"))
	})

	r.Route("/invitations", func(r chi.Router) {
		r.Get("/{id}", invitationHandler.GetInvitationCompleteData)
		r.Get("/{id}/group/{waNumber}", invitationHandler.GetInvitationCompleteDataByWANumber)
		r.Post("/", invitationHandler.CreateInvitation)
	})

	r.Route("/users", func(r chi.Router) {
		r.Post("/{id}", userHandler.CreateUser)
		r.Put("/{id}", userHandler.UpdateUser)
		r.Post("/{id}/comment", userHandler.CreateUserComment)
		r.Put("/{id}/comment", userHandler.UpdateUserComment)
		r.Put("/{id}/comment/like", userHandler.LikeUnlikeComment)
		r.Get("/{id}/comments", userHandler.GetUserCommentList)
		r.Get("/{id}/comments/specific", userHandler.GetUserSpecificComment)
		r.Post("/{id}/rsvp", userHandler.CreateUserRSVP)
		r.Post("/{id}/reminder/date", userHandler.RemindUserWeddingDate)
		r.Post("/{id}/reminder/video", userHandler.RemindUserSendWeddingVideo)
		r.Get("/{id}/qr-code/download", userHandler.DownloadQRCode)

		r.Get("/current-video", userHandler.GetCurrentVideo)
		r.Get("/qr-rsvp/{invitation_code}", userHandler.ValidateUserQRRsvp)
		r.Post("/synchronize-data", userHandler.SynchronizeUser)
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/", token.HandleMain)
	})

	// STATIC FILE SERVE (FOR DEVELOPMENT PURPOSE ONLY)
	fs := http.FileServer(http.Dir("static/qr-codes"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	return r
}
