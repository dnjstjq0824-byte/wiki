package search

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/dnjstjqppk-dev/treewiki/pkg/validate"
	"github.com/dnjstjqppk-dev/treewiki/services/identity"
)

type Handler struct {
	index SearchBackend
	media MediaStorage
}

func NewHandler(index SearchBackend, media MediaStorage) *Handler {
	return &Handler{index: index, media: media}
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	ns := -1
	if n := r.URL.Query().Get("namespace"); n != "" {
		// parse int
		fmt.Sscan(n, &ns)
	}
	category := r.URL.Query().Get("category")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	res := h.index.SearchFaceted(q, ns, category, from, to)
	if res == nil {
		res = []SearchResult{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(res, validate.ParsePagination(r, 20)))
		return
	}
	h.writeJSON(w, http.StatusOK, res)
}

func (h *Handler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		q = r.URL.Query().Get("prefix")
	}
	limit := 10
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n <= 50 {
			limit = n
		}
	}
	h.writeJSON(w, http.StatusOK, h.index.Autocomplete(q, limit))
}

func (h *Handler) ListMedia(w http.ResponseWriter, r *http.Request) {
	assets := h.media.List()
	if assets == nil {
		assets = []*MediaAsset{}
	}
	if validate.HasPagination(r) {
		h.writeJSON(w, http.StatusOK, validate.Paginate(assets, validate.ParsePagination(r, 20)))
		return
	}
	h.writeJSON(w, http.StatusOK, assets)
}
func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid multipart form"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file required"})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "read file failed"})
		return
	}

	hash := sha256.Sum256(data)
	sha := hex.EncodeToString(hash[:])

	// Sanitize the client filename (traversal defense) and cap its length.
	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." || filename == "/" {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid filename"})
		return
	}
	if len([]rune(filename)) > 255 {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "filename too long (max 255 characters)"})
		return
	}
	_ = os.MkdirAll("uploads", 0755)
	dest := filepath.Join("uploads", filename)
	if err := os.WriteFile(dest, data, 0644); err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "store file failed"})
		return
	}

	// Authenticated identity wins over the spoofable form field.
	uploaderID := r.FormValue("uploader_id")
	if ctx := identity.GetUserContext(r); ctx.UserID != "" {
		uploaderID = ctx.UserID
	}
	asset := &MediaAsset{
		Filename:     filename,
		OriginalName: filename,
		MimeType:     header.Header.Get("Content-Type"),
		FileSize:     int64(len(data)),
		S3Key:        dest,
		SHA256:       sha,
		LicenseType:  r.FormValue("license_type"),
		UploaderID:   uploaderID,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	if asset.LicenseType == "" {
		asset.LicenseType = "CC_BY_SA_4_0"
	}
	if !ValidateLicense(asset.LicenseType) {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid license type"})
		return
	}
	if err := h.media.Add(asset); err != nil {
		h.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	h.writeJSON(w, http.StatusCreated, asset)
}

func ValidateLicense(license string) bool {
	allowed := []string{"CC_BY_SA_4_0", "PUBLIC_DOMAIN", "FAIR_USE"}
	for _, a := range allowed {
		if license == a {
			return true
		}
	}
	return false
}

func SanitizeSVG(input string) string {
	// Remove script tags and on* attributes (basic)
	re := regexp.MustCompile(`(?i)<script[\s\S]*?</script>`)
	out := re.ReplaceAllString(input, "")
	// remove on* attributes
	re2 := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*"[^"]*"`)
	out = re2.ReplaceAllString(out, "")
	return out
}
