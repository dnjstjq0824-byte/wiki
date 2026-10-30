package watchlist

import (
	"encoding/json"
	"net/http"
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
	h.service.Add(body.UserID, body.PageKey, body.NotifyEmail)
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
	h.service.Remove(body.UserID, body.PageKey)
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "unwatched"})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	h.writeJSON(w, http.StatusOK, h.service.List(userID))
}