package wiki

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/dnjstjqppk-dev/treewiki/pkg/validate"
	"github.com/dnjstjqppk-dev/treewiki/services/identity"
)

type PageSaveCallback func(page *Page, isUpdate bool)

type Handler struct {
	service    *WikiService
	onPageSave PageSaveCallback
}

func NewHandler(service *WikiService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetOnPageSave(cb PageSaveCallback) {
	h.onPageSave = cb
}

func parseNamespaceParam(s string) Namespace {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "Main") || s == "0" {
		return NSMain
	}
	if n, ok := ParseNamespace(s); ok {
		return n
	}
	// try canonical names
	for _, info := range AllNamespaces() {
		if strings.EqualFold(info.Name, s) || strings.EqualFold(info.CanonicalName, s) {
			return Namespace(info.ID)
		}
	}
	return NSMain
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) writeErr(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

func (h *Handler) GetPage(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	page, err := h.service.GetPage(ns, title)
	if err != nil {
		if err == ErrPageNotFound {
			h.writeErr(w, http.StatusNotFound, err.Error())
			return
		}
		if err == ErrPageDeleted {
			h.writeErr(w, http.StatusGone, err.Error())
			return
		}
		h.writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, page)
}
func (h *Handler) Namespaces(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, AllNamespaces())
}

func (h *Handler) PutPage(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}

	ctx := identity.GetUserContext(r)
	authorID := ctx.Username
	if authorID == "" {
		authorID = ctx.UserID
	}
	if authorID == "" {
		authorID = ctx.MaskedIP
	}
	if authorID == "" {
		authorID = "anonymous"
	}
	userRole := ctx.Role

	var body struct {
		Wikitext    string `json:"wikitext"`
		Summary     string `json:"summary"`
		IsMinor     bool   `json:"is_minor"`
		ParentRevID string `json:"parent_revision_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}

	// check existence
	existing, err := h.service.GetPage(ns, title)
	expectedRev := r.Header.Get("If-Match")
	if expectedRev == "" {
		expectedRev = body.ParentRevID
	}

	if err == ErrPageNotFound {
		// create
		page, err := h.service.CreatePage(ns, title, body.Wikitext, body.Summary, authorID, userRole)
		if err != nil {
			h.writeErr(w, http.StatusBadRequest, err.Error())
			return
		}
		if h.onPageSave != nil {
			h.onPageSave(page, false)
		}
		h.writeJSON(w, http.StatusCreated, page)
		return
	}
	if err != nil {
		h.writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = existing

	page, err := h.service.UpdatePage(ns, title, body.Wikitext, body.Summary, authorID, userRole, expectedRev, body.IsMinor)
	if err != nil {
		if err == ErrRevisionConflict {
			h.writeErr(w, http.StatusConflict, err.Error())
			return
		}
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if h.onPageSave != nil {
		h.onPageSave(page, true)
	}
	h.writeJSON(w, http.StatusOK, page)
}

func (h *Handler) DeletePage(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	ctx := identity.GetUserContext(r)
	if err := h.service.DeletePage(ns, title, ctx.Role); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) UndeletePage(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	ctx := identity.GetUserContext(r)
	if err := h.service.UndeletePage(ns, title, ctx.Role); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "undeleted"})
}
func (h *Handler) MovePage(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	ctx := identity.GetUserContext(r)
	var body struct {
		NewNamespace   string `json:"new_namespace"`
		NewTitle       string `json:"new_title"`
		CreateRedirect bool   `json:"create_redirect"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	newNS := parseNamespaceParam(body.NewNamespace)
	if err := h.service.MovePage(ns, title, newNS, body.NewTitle, ctx.UserID, ctx.Role, body.CreateRedirect); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "moved"})
}

func (h *Handler) Backlinks(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	pages := h.service.GetBacklinks(ns, title)
	h.writeJSON(w, http.StatusOK, pages)
}

func (h *Handler) CategoryMembers(w http.ResponseWriter, r *http.Request) {
	cat := r.PathValue("category")
	if cat == "" {
		cat = r.PathValue("name")
	}
	pages := h.service.GetCategoryMembers(cat)
	h.writeJSON(w, http.StatusOK, pages)
}

func (h *Handler) Revisions(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	revs := h.service.GetRevisions(ns, title)
	if revs == nil {
		revs = []*Revision{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(revs, validate.ParsePagination(r, 50)))
		return
	}
	h.writeJSON(w, http.StatusOK, revs)
}

func (h *Handler) SetProtection(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	ctx := identity.GetUserContext(r)
	var body struct {
		Level string `json:"level"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if err := h.service.SetProtection(ns, title, body.Level, ctx.Role); err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, map[string]string{"status": "protection updated"})
}

func (h *Handler) ListPages(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	pages := h.service.ListPages(ns)
	if pages == nil {
		pages = []*Page{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(pages, validate.ParsePagination(r, 50)))
		return
	}
	h.writeJSON(w, http.StatusOK, pages)
}

func (h *Handler) ResolvePage(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	page, err := h.service.ResolvePage(ns, title)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, page)
}

func (h *Handler) Revert(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	ctx := identity.GetUserContext(r)
	var body struct {
		RevisionID string `json:"revision_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	page, err := h.service.RevertToRevision(ns, title, body.RevisionID, ctx.UserID, ctx.Role)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, page)
}

func (h *Handler) Rollback(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	ctx := identity.GetUserContext(r)
	page, err := h.service.FastRollback(ns, title, ctx.UserID, ctx.Role)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, page)
}

func (h *Handler) RandomPage(w http.ResponseWriter, r *http.Request) {
	var ns *Namespace
	if raw := r.URL.Query().Get("namespace"); raw != "" {
		parsed := parseNamespaceParam(raw)
		ns = &parsed
	}
	page, err := h.service.RandomPage(ns)
	if err != nil {
		h.writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, page)
}

func (h *Handler) Diff(w http.ResponseWriter, r *http.Request) {
	ns := parseNamespaceParam(r.PathValue("namespace"))
	title := r.PathValue("slug")
	if title == "" {
		title = r.PathValue("title")
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	chunks, err := h.service.DiffRevisions(ns, title, from, to)
	if err != nil {
		h.writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	h.writeJSON(w, http.StatusOK, chunks)
}
