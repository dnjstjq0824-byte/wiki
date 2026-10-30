package talk

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dnjstjqppk-dev/treewiki/services/identity"
)

type Handler struct {
	service *TalkService
}

func NewHandler(service *TalkService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeErr(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) CreateThread(w http.ResponseWriter, r *http.Request) {
	var body struct {
		PageKey  string `json:"page_key"`
		Title    string `json:"title"`
		AuthorID string `json:"author_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.AuthorID == "" {
		ctx := identity.GetUserContext(r)
		if ctx.Username != "" {
			body.AuthorID = ctx.Username
		} else if ctx.UserID != "" {
			body.AuthorID = ctx.UserID
		} else if ctx.MaskedIP != "" {
			body.AuthorID = ctx.MaskedIP
		} else {
			body.AuthorID = "anonymous"
		}
	}
	t, err := h.service.CreateThread(body.PageKey, body.Title, body.AuthorID)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) ListThreads(w http.ResponseWriter, r *http.Request) {
	pageKey := r.URL.Query().Get("page_key")
	h.writeJSON(w, http.StatusOK, h.service.ListThreads(pageKey))
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status TalkStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.SetStatus(id, body.Status); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	var body struct {
		ParentID string `json:"parent_id"`
		AuthorID string `json:"author_id"`
		Content  string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if body.AuthorID == "" {
		ctx := identity.GetUserContext(r)
		if ctx.Username != "" {
			body.AuthorID = ctx.Username
		} else if ctx.UserID != "" {
			body.AuthorID = ctx.UserID
		} else if ctx.MaskedIP != "" {
			body.AuthorID = ctx.MaskedIP
		} else {
			body.AuthorID = "anonymous"
		}
	}
	c, err := h.service.AddComment(threadID, body.ParentID, body.AuthorID, body.Content)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, c)
}

func (h *Handler) GetComments(w http.ResponseWriter, r *http.Request) {
	threadID := r.PathValue("id")
	h.writeJSON(w, http.StatusOK, h.service.GetComments(threadID))
}

func (h *Handler) CreateRFC(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ThreadID string `json:"thread_id"`
		Question string `json:"question"`
		Quorum   int    `json:"quorum"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	rfc, err := h.service.CreateRFC(body.ThreadID, body.Question, body.Quorum)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, rfc)
}

func (h *Handler) VoteRFC(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		UserID string `json:"user_id"`
		Choice string `json:"choice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.VoteRFC(id, body.UserID, body.Choice); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "voted"})
}

func (h *Handler) GetRFC(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	rfc, err := h.service.GetRFC(id)
	if err != nil {
		h.writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, rfc)
}

func (h *Handler) CloseRFC(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.CloseRFC(id); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

func (h *Handler) SSE(w http.ResponseWriter, r *http.Request) {
	pageKey := r.URL.Query().Get("page_key")
	if pageKey == "" {
		h.writeErr(w, http.StatusBadRequest, "page_key required")
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeErr(w, http.StatusInternalServerError, "streaming not supported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch, cancel := h.service.Subscribe(pageKey)
	defer cancel()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}
func (h *Handler) ArchiveThread(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.ArchiveThread(id); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (h *Handler) ListArchived(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.ListArchived())
}

func (h *Handler) CreateArbitration(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ThreadID string `json:"thread_id"`
		Summary  string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	a, err := h.service.CreateArbitration(body.ThreadID, body.Summary)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) ListArbitrations(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.ListArbitrations())
}
func (h *Handler) PinThread(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Pinned bool `json:"pinned"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.PinThread(id, body.Pinned); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) LockThread(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Locked bool `json:"locked"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.LockThread(id, body.Locked); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}