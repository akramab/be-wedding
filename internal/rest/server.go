package rest

import (
	"be-wedding/internal/config"
	invitationhandler "be-wedding/internal/rest/handler/invitation"
	userhandler "be-wedding/internal/rest/handler/user"
	"be-wedding/internal/rest/middleware"
	storepgsql "be-wedding/internal/store/pgsql"
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
	whatsAppClient whatsapp.Client,

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

	invitationHandler := invitationhandler.NewInvitationHandler(cfg.API, sqlDB, invitationStore)
	userHandler := userhandler.NewUserHandler(cfg.API, sqlDB, userStore, invitationStore, whatsAppClient)

	r.Route("/invitations", func(r chi.Router) {
		r.Get("/{id}", invitationHandler.GetInvitationCompleteData)
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
	})

	r.Route("/auth", func(r chi.Router) {
		r.Get("/", token.HandleMain)
	})

	// STATIC FILE SERVE (FOR DEVELOPMENT PURPOSE ONLY)
	fs := http.FileServer(http.Dir("static/qr-codes"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	return r
}
