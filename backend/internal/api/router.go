package api

import (
	"net/http"
	"xpired/internal/auth"
	database "xpired/internal/db"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func SetupRoutes(
	db *database.DB,
) http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://campus-connect-liard-ten.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Cookie"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	repo := database.NewRepository(db)
	handler := NewHandler(repo)

	r.Get("/health", handler.HealthHandler)

	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", handler.RegisterHandler)
			r.Post("/signin", handler.LoginHandler)

			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Get("/me", handler.UserProfileHandler)
				r.Post("/logout", handler.LogoutHandler)
			})
		})

		r.Route("/documents", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(auth.AuthMiddleware)
				r.Get("/", handler.ListDocumentsHandler)
				r.Post("/", handler.CreateDocumentHandler)
				r.Get("/{id}", handler.GetDocumentHandler)
				r.Put("/{id}", handler.UpdateDocumentHandler)
				r.Delete("/{id}", handler.DeleteDocumentHandler)
				r.Get("/{id}/reminders", handler.GetDocumentRemindersHandler)
				r.Put("/{id}/reminders", handler.ToggleDocumentReminderHandler)
			})
		})

		r.Get("/reminder-intervals", handler.GetReminderIntervalsHandler)
	})

	return r
}
