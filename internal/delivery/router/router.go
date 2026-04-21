package router

import (
	"ZVideo/internal/delivery/handlers"
	"ZVideo/internal/delivery/middleware"
	"ZVideo/internal/domain"
	"ZVideo/internal/service"

	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	Auth               *handlers.AuthHandler
	Channel            *handlers.ChannelHandler
	Video              *handlers.VideoHandler
	Subscription       *handlers.SubscriptionHandler
	VideoInteraction   *handlers.VideoInteractionHandler
	Comment            *handlers.CommentHandler
	CommentInteraction *handlers.CommentInteractionHandler
	Admin              *handlers.AdminHandler
	Playlist           *handlers.PlaylistHandler
}

func NewRouter(h *Handlers, authSvc service.AuthService, baseLogger domain.Logger) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Recovery)
	r.Use(middleware.Logging(baseLogger))

	r.Post("/register", h.Auth.Register)
	r.Post("/login", h.Auth.Login)
	r.Post("/refresh", h.Auth.Refresh)
	r.Post("/logout", h.Auth.Logout)
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc))

		r.Get("/me", h.Auth.GetMe)
		r.Patch("/me/notifications", h.Auth.UpdateNotificationsSettings)

		r.Get("/channels/{id}", h.Channel.GetChannel)
		r.Get("/channels/me", h.Channel.GetMyChannel)
		r.Post("/channels", h.Channel.CreateChannel)
		r.Patch("/channels/{id}", h.Channel.UpdateChannel)
		r.Delete("/channels/{id}", h.Channel.DeleteChannel)

		r.Post("/videos", h.Video.CreateVideo)
		r.Get("/videos/{id}", h.Video.GetVideo)
		r.Patch("/videos/{id}", h.Video.UpdateVideo)
		r.Delete("/videos/{id}", h.Video.DeleteVideo)
		r.Get("/videos", h.Video.List)
		r.Get("/videos/me", h.Video.ListMyVideos)
		r.Get("/channels/{channelID}/videos", h.Video.ListChannelVideos)
		r.Post("/videos/upload-url", h.Video.GetUploadPresignedURL)
		r.Get("/videos/{id}/stream-url", h.Video.GetStreamingPresignedURL)

		r.Post("/videos/{id}/like", h.VideoInteraction.Like)
		r.Post("/videos/{id}/dislike", h.VideoInteraction.Dislike)
		r.Delete("/videos/{id}/rating", h.VideoInteraction.RemoveRating)

		r.Post("/channels/{id}/subscribe", h.Subscription.Subscribe)
		r.Delete("/channels/{id}/subscribe", h.Subscription.Unsubscribe)
		r.Get("/subscriptions", h.Subscription.GetUserSubscriptions)
		r.Get("/channels/{channelID}/playlists", h.Playlist.ListByChannel)
		r.Post("/channels/{channelID}/playlists", h.Playlist.Create)
		r.Get("/playlists/{id}", h.Playlist.GetByID)
		r.Patch("/playlists/{id}", h.Playlist.Update)
		r.Delete("/playlists/{id}", h.Playlist.Delete)
		r.Post("/playlists/{id}/videos/{videoID}", h.Playlist.AddVideo)
		r.Delete("/playlists/{id}/videos/{videoID}", h.Playlist.RemoveVideo)

		r.Post("/videos/{videoID}/comments", h.Comment.Create)
		r.Get("/videos/{videoID}/comments", h.Comment.List)
		r.Patch("/comments/{id}", h.Comment.Update)
		r.Delete("/comments/{id}", h.Comment.Delete)

		r.Post("/comments/{id}/like", h.CommentInteraction.Like)
		r.Post("/comments/{id}/dislike", h.CommentInteraction.Dislike)
		r.Delete("/comments/{id}/rating", h.CommentInteraction.RemoveRating)

		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireRole("admin"))
			r.Post("/admin/users/{id}/ban", h.Admin.BanUser)
			r.Post("/admin/users/{id}/unban", h.Admin.UnbanUser)
		})
	})

	return r
}
