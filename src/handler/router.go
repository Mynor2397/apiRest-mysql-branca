package handler

import (
	"net/http"

	"github.com/matryer/way"

	"github.com/Mynor2397/social-network/src/service"
)

type handler struct {
	*service.Service
}

//New crea un handler con ruteo predefinido
func New(s *service.Service) http.Handler {
	api := way.NewRouter()
	h := &handler{s}

	api.HandleFunc("POST", "/login", h.login)
	api.HandleFunc("POST", "/users", h.createUser)
	api.HandleFunc("GET", "/auth_user", h.authUser)
	api.HandleFunc("GET", "/users", h.users)
	api.HandleFunc("GET", "/users/:username", h.user)
	api.HandleFunc("POST", "/users/:username/toggle_follow", h.toggleFollow)

	r := way.NewRouter()
	r.Handle("*", "/api...", http.StripPrefix("/api", h.withAuth(api)))

	return r
}
