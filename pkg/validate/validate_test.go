package validate

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParsePaginationDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	p := ParsePagination(req, 20)
	if p.Page != 1 || p.PerPage != 20 {
		t.Fatalf("expected 1/20, got %d/%d", p.Page, p.PerPage)
	}
}

func TestParsePaginationClamps(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x?page=0&per_page=500", nil)
	p := ParsePagination(req, 20)
	if p.Page != 1 || p.PerPage != 100 {
		t.Fatalf("expected clamp to 1/100, got %d/%d", p.Page, p.PerPage)
	}
}

func TestParsePaginationIgnoresGarbage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x?page=abc&per_page=-3", nil)
	p := ParsePagination(req, 25)
	if p.Page != 1 || p.PerPage != 25 {
		t.Fatalf("expected defaults 1/25, got %d/%d", p.Page, p.PerPage)
	}
}

func TestPaginateSlices(t *testing.T) {
	items := []int{1, 2, 3, 4, 5}
	got := Paginate(items, Pagination{Page: 2, PerPage: 2})
	if got.Total != 5 || len(got.Data) != 2 || got.Data[0] != 3 {
		t.Fatalf("unexpected page: %+v", got)
	}
}

func TestPaginateBeyondEnd(t *testing.T) {
	got := Paginate([]int{1}, Pagination{Page: 9, PerPage: 10})
	if got.Total != 1 || len(got.Data) != 0 {
		t.Fatalf("expected empty page with total 1: %+v", got)
	}
}

func TestPaginateNilYieldsEmptyArray(t *testing.T) {
	got := Paginate[int](nil, Pagination{Page: 1, PerPage: 10})
	if got.Data == nil || len(got.Data) != 0 {
		t.Fatalf("expected non-nil empty data: %+v", got)
	}
}

func TestFieldValidators(t *testing.T) {
	if err := NonEmpty("title", "  "); err == nil {
		t.Fatal("expected NonEmpty to reject blank")
	}
	if err := NonEmpty("title", "x"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := MaxLen("summary", "abcdef", 3); err == nil {
		t.Fatal("expected MaxLen to reject")
	}
	if err := OneOf("action", "NOPE", "TAG", "WARN"); err == nil {
		t.Fatal("expected OneOf to reject")
	}
	if err := OneOf("action", "TAG", "TAG", "WARN"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
