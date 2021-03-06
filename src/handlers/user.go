package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/matryer/way"

	"github.com/Mynor2397/social-network/src/service"
)

type createUserInput struct {
	Email, Username, Password string
}

func (h *handler) createUser(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// w.Header().Set("Access-Control-Allow-Origin", "*")

	var in createUserInput
	defer r.Body.Close()

	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := h.CreateUser(r.Context(), in.Email, in.Username, in.Password)
	if err == service.ErrInvalideEmail || err == service.ErrInvalideUsername || err == service.ErrInvalidUser {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrInvalidUser {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err == service.ErrInvalidPassword {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	if err == service.ErrUserOk {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"text":"Usuario ingresado correctamente!"}`))
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) user(w http.ResponseWriter, r *http.Request) {

	ctx := r.Context()
	username := way.Param(ctx, "username")
	u, err := h.User(ctx, username)
	if err == service.ErrInvalideUsername {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
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

func (h *handler) toggleFollow(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	username := way.Param(ctx, "username")

	out, err := h.ToggleFollow(ctx, username)
	if err == service.ErrUnauthenticated {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if err == service.ErrInvalideUsername {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if err == service.ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if err == service.ErrForbiddenFollow {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	if err != nil {
		respondError(w, err)
		return
	}

	respond(w, out, http.StatusOK)
}

func (h *handler) users(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	search := q.Get("search")
	first, _ := strconv.Atoi(q.Get("first"))
	after := q.Get("after")
	uu, err := h.Users(r.Context(), search, first, after)
	if err != nil {
		respondError(w, err)
		return
	}
	respond(w, uu, http.StatusOK)
}
