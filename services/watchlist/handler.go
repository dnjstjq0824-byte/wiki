package watchlist

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dnjstjqppk-dev/treewiki/pkg/validate"
	"github.com/dnjstjqppk-dev/treewiki/services/identity"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID      string `json:"user_id"`
		PageKey     string `json:"page_key"`
		NotifyEmail bool   `json:"notify_email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if body.UserID == "" {
		body.UserID = identity.GetUserContext(r).UserID
	}
	if body.UserID == "" || body.PageKey == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id and page_key are required"})
		return
	}
	if err := h.service.Add(body.UserID, body.PageKey, body.NotifyEmail); err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "watching"})
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	var body struct {
		UserID  string `json:"user_id"`
		PageKey string `json:"page_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}
	if body.UserID == "" {
		body.UserID = identity.GetUserContext(r).UserID
	}
	if body.UserID == "" || body.PageKey == "" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "user_id and page_key are required"})
		return
	}
	if err := h.service.Remove(body.UserID, body.PageKey); err != nil {
		h.writeErr(w, err)
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "unwatched"})
}

// writeErr maps store errors to 404 (unknown page/user) or 400.
func (h *Handler) writeErr(w http.ResponseWriter, err error) {
	msg := err.Error()
	low := strings.ToLower(msg)
	if strings.Contains(low, "not found") || strings.Contains(low, "unknown") {
		h.writeJSON(w, http.StatusNotFound, map[string]string{"error": msg})
		return
	}
	h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = identity.GetUserContext(r).UserID
	}
	items := h.service.List(userID)
	if items == nil {
		items = []*WatchEntry{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(items, validate.ParsePagination(r, 50)))
		return
	}
	h.writeJSON(w, http.StatusOK, items)
}
