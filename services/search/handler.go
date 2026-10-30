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
)

type Handler struct {
	index *SearchIndex
	media *MediaStore
}

func NewHandler(index *SearchIndex, media *MediaStore) *Handler {
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
	h.writeJSON(w, http.StatusOK, res)
}

func (h *Handler) Autocomplete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	h.writeJSON(w, http.StatusOK, h.index.Autocomplete(q, 10))
}

func (h *Handler) ListMedia(w http.ResponseWriter, r *http.Request) {
	h.writeJSON(w, http.StatusOK, h.media.List())
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

	_ = os.MkdirAll("uploads", 0755)
	dest := filepath.Join("uploads", header.Filename)
	_ = os.WriteFile(dest, data, 0644)

	asset := &MediaAsset{
		Filename:     header.Filename,
		OriginalName: header.Filename,
		MimeType:     header.Header.Get("Content-Type"),
		FileSize:     int64(len(data)),
		S3Key:        dest,
		SHA256:       sha,
		LicenseType:  r.FormValue("license_type"),
		UploaderID:   r.FormValue("uploader_id"),
		CreatedAt:    "now",
	}
	if asset.LicenseType == "" {
		asset.LicenseType = "CC_BY_SA_4_0"
	}
	if !ValidateLicense(asset.LicenseType) {
		h.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid license type"})
		return
	}
	h.media.Add(asset)
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