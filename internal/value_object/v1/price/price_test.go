package price

import (
	"testing"
)

func TestPriceNewAndNull(t *testing.T) {
	p, err := NewPrice("150000.00")
	if err != nil {
		t.Fatalf("NewPrice failed: %v", err)
	}
	if p.IsNull() {
		t.Fatal("expected non-null price")
	}
	if p.String() != "150000.00" {
		t.Errorf("expected '150000.00', got %q", p.String())
	}

	np := NullPrice()
	if !np.IsNull() {
		t.Fatal("expected null price")
	}
}

func TestPriceNonNegative(t *testing.T) {
	_, err := NewPrice("-100.00")
	if err == nil {
		t.Fatal("expected error for negative price")
	}
}

func TestPriceInvalid(t *testing.T) {
	_, err := NewPrice("abc")
	if err == nil {
		t.Fatal("expected error for invalid price")
	}
	_, err = NewPrice("")
	if err != nil {
		t.Fatalf("empty price should be null, got err: %v", err)
	}
}

func TestPriceJSON(t *testing.T) {
	p, _ := NewPrice("150000.00")
	b, err := p.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "150000.00" {
		t.Errorf("expected '150000.00', got %s", b)
	}

	np := NullPrice()
	b, _ = np.MarshalJSON()
	if string(b) != "null" {
		t.Errorf("expected 'null', got %s", b)
	}

	var p2 Price
	if err := p2.UnmarshalJSON([]byte("2500.50")); err != nil {
		t.Fatal(err)
	}
	if p2.IsNull() || p2.String() != "2500.50" {
		t.Errorf("expected 2500.50, got %v", p2.String())
	}

	var p3 Price
	if err := p3.UnmarshalJSON([]byte("null")); err != nil {
		t.Fatal(err)
	}
	if !p3.IsNull() {
		t.Fatal("expected null after unmarshaling null")
	}

	var p4 Price
	if err := p4.UnmarshalJSON([]byte("-5.00")); err == nil {
		t.Fatal("expected error for negative price in JSON")
	}
}

func TestPriceFloat64(t *testing.T) {
	p, _ := NewPrice("150000.00")
	if p.Float64() != 150000.0 {
		t.Errorf("expected 150000.0, got %f", p.Float64())
	}
}
