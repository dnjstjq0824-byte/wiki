package topictree

import (
	"encoding/json"
	"net/http"

	"github.com/dnjstjqppk-dev/treewiki/pkg/validate"
)

type Handler struct {
	service *TopicService
}

func NewHandler(service *TopicService) *Handler {
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

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var n TopicNode
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.AddNode(&n); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, n)
}

func (h *Handler) GetNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	n, err := h.service.GetNode(id)
	if err != nil {
		h.writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, n)
}

func (h *Handler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var n TopicNode
	if err := json.NewDecoder(r.Body).Decode(&n); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	n.ID = id
	if err := h.service.UpdateNode(&n); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, n)
}

func (h *Handler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.service.DeleteNode(id); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) Link(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ParentID string `json:"parent_id"`
		ChildID  string `json:"child_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.Link(body.ParentID, body.ChildID); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "linked"})
}

func (h *Handler) Unlink(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ParentID string `json:"parent_id"`
		ChildID  string `json:"child_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.Unlink(body.ParentID, body.ChildID); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "unlinked"})
}

func (h *Handler) Ancestors(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.writeJSON(w, http.StatusOK, h.service.Ancestors(id))
}

func (h *Handler) Descendants(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	h.writeJSON(w, http.StatusOK, h.service.Descendants(id))
}

func (h *Handler) CreatePath(w http.ResponseWriter, r *http.Request) {
	var p LearningPath
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.CreatePath(&p); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, p)
}

func (h *Handler) GetPath(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.service.GetPath(id)
	if err != nil {
		h.writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, p)
}

func (h *Handler) ListPaths(w http.ResponseWriter, r *http.Request) {
	items := h.service.AllPaths()
	if items == nil {
		items = []*LearningPath{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(items, validate.ParsePagination(r, 50)))
		return
	}
	h.writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ListNodes(w http.ResponseWriter, r *http.Request) {
	items := h.service.AllNodes()
	if items == nil {
		items = []*TopicNode{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(items, validate.ParsePagination(r, 50)))
		return
	}
	h.writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.Export())
}
