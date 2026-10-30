package governance

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	service *GovernanceService
}

func NewHandler(service *GovernanceService) *Handler {
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

func (h *Handler) AddFilter(w http.ResponseWriter, r *http.Request) {
	var f AbuseFilter
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.AddFilter(&f); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, f)
}

func (h *Handler) ListFilters(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.ListFilters())
}

func (h *Handler) Evaluate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Text      string `json:"text"`
		PageTitle string `json:"page_title"`
		UserID    string `json:"user_id"`
		IP        string `json:"ip"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	action, f := h.service.Evaluate(body.Text, body.PageTitle, body.UserID, body.IP)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{"action": action, "filter": f})
}

func (h *Handler) AddBlock(w http.ResponseWriter, r *http.Request) {
	var b Block
	if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.AddBlock(&b); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) ListBlocks(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.ListBlocks())
}

func (h *Handler) AddOversight(w http.ResponseWriter, r *http.Request) {
	var o OversightLog
	if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	h.service.AddOversight(&o)
	h.writeJSON(w, http.StatusCreated, o)
}

func (h *Handler) ListOversight(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.ListOversight())
}

func (h *Handler) Dump(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"filters":    h.service.ListFilters(),
		"blocks":     h.service.ListBlocks(),
		"oversight":  h.service.ListOversight(),
		"takedowns":  h.service.ListTakedowns(),
		"recent":     h.service.RecentChanges(),
	})
}
func (h *Handler) AddTakedown(w http.ResponseWriter, r *http.Request) {
	var t TakedownRequest
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.AddTakedown(&t); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusCreated, t)
}

func (h *Handler) ListTakedowns(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.ListTakedowns())
}

func (h *Handler) UpdateTakedownStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.UpdateTakedownStatus(id, body.Status); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}
func (h *Handler) RecentChanges(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.service.RecentChanges())
}
