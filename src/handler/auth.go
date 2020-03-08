package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Mynor2397/social-network/src/service"
)

type loginInput struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
}

func (h *handler) login(w http.ResponseWriter, r *http.Request) {
	var in loginInput
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	out, err := h.Login(r.Context(), in.Email, in.Password)
	if err == service.ErrInvalideEmail {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err == service.ErrInvalidPassword {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, out, http.StatusOK)
}

func (h *handler) withAuth(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authentication, Content-Length, Accept-Encoding, X-CSRF-Token")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		a := r.Header.Get("Authorization")

		if !strings.HasPrefix(a, "Bearer") {
			next.ServeHTTP(w, r)
			return
		}

		token := a[7:]
		uid, err := h.AuthUserID(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, service.KeyAuthUser, uid)
		next.ServeHTTP(w, r.WithContext(ctx))

	})

}

func (h *handler) authUser(w http.ResponseWriter, r *http.Request) {

	u, err := h.AuthUser(r.Context())

	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, u, http.StatusOK)

}
