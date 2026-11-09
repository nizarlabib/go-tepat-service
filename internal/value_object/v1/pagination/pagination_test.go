package pagination

import (
	"net/url"
	"testing"
)

func TestNewPaginationDefaultValues(t *testing.T) {
	pg := NewPagination(url.Values{})
	if pg.GetPage() != 1 {
		t.Errorf("expected default page 1, got %d", pg.GetPage())
	}
	if pg.GetLimit() != 10 {
		t.Errorf("expected default limit 10, got %d", pg.GetLimit())
	}
}

func TestNewPaginationCustomValues(t *testing.T) {
	query := url.Values{}
	query.Set("page", "3")
	query.Set("limit", "25")
	pg := NewPagination(query)
	if pg.GetPage() != 3 {
		t.Errorf("expected page 3, got %d", pg.GetPage())
	}
	if pg.GetLimit() != 25 {
		t.Errorf("expected limit 25, got %d", pg.GetLimit())
	}
}

func TestNewPaginationLimitCap(t *testing.T) {
	query := url.Values{}
	query.Set("limit", "200")
	pg := NewPagination(query)
	if pg.GetLimit() > 50 {
		t.Errorf("expected limit capped at 50, got %d", pg.GetLimit())
	}
	if pg.GetLimit() != 50 {
		t.Errorf("expected limit exactly 50, got %d", pg.GetLimit())
	}
}

func TestNewPaginationInvalidValues(t *testing.T) {
	query := url.Values{}
	query.Set("page", "invalid")
	query.Set("limit", "invalid")
	pg := NewPagination(query)
	if pg.GetPage() != 1 {
		t.Errorf("expected default page 1 for invalid input, got %d", pg.GetPage())
	}
	if pg.GetLimit() != 10 {
		t.Errorf("expected default limit 10 for invalid input, got %d", pg.GetLimit())
	}
}

func TestGetOffset(t *testing.T) {
	pg := &Pagination{Page: 3, Limit: 10}
	if pg.GetOffset() != 20 {
		t.Errorf("expected offset 20, got %d", pg.GetOffset())
	}
}

func TestGetLimitCapConsistency(t *testing.T) {
	// GetLimit should also cap at 50, consistent with NewPagination
	pg := &Pagination{Limit: 200}
	if pg.GetLimit() != 50 {
		t.Errorf("expected GetLimit to cap at 50, got %d", pg.GetLimit())
	}
}
