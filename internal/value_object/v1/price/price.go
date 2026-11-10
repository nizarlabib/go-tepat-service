package price

import (
	"database/sql/driver"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// Price represents a monetary value in IDR with NUMERIC(15,2) precision.
// It is stored as a string to preserve exact decimal precision and avoids
// float64/float32 for money. Zero value is a null price.
type Price struct {
	value string
	valid bool
}

// NullPrice returns a null Price.
func NullPrice() Price {
	return Price{}
}

// NewPrice creates a Price from a numeric string (e.g. "150000.00").
// Returns error if the string is not a valid non-negative decimal.
func NewPrice(s string) (Price, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Price{}, nil
	}

	rat, ok := new(big.Rat).SetString(s)
	if !ok {
		return Price{}, fmt.Errorf("invalid price value: %q", s)
	}

	if rat.Sign() < 0 {
		return Price{}, fmt.Errorf("price must be non-negative: %q", s)
	}

	return Price{value: normalizeToTwoDecimals(rat), valid: true}, nil
}

// NewPriceFromBigRat creates a Price from a big.Rat.
func NewPriceFromBigRat(rat *big.Rat) (Price, error) {
	if rat == nil {
		return Price{}, nil
	}
	if rat.Sign() < 0 {
		return Price{}, fmt.Errorf("price must be non-negative")
	}
	return Price{value: normalizeToTwoDecimals(rat), valid: true}, nil
}

// MustNewPrice is like NewPrice but panics on error.
func MustNewPrice(s string) Price {
	p, err := NewPrice(s)
	if err != nil {
		panic(err)
	}
	return p
}

// IsNull reports whether the price is null.
func (p Price) IsNull() bool {
	return !p.valid
}

// String returns the string representation of the price.
func (p Price) String() string {
	if !p.valid {
		return ""
	}
	return p.value
}

// Float64 returns the price as float64 (may lose precision; use for display only).
func (p Price) Float64() float64 {
	if !p.valid {
		return 0
	}
	f, _ := new(big.Float).SetString(p.value)
	v, _ := f.Float64()
	return v
}

// BigRat returns the price as big.Rat.
func (p Price) BigRat() *big.Rat {
	if !p.valid {
		return nil
	}
	v, _ := new(big.Rat).SetString(p.value)
	return v
}

// Value implements driver.Valuer for storing into NUMERIC column.
func (p Price) Value() (driver.Value, error) {
	if !p.valid {
		return nil, nil
	}
	return p.value, nil
}

// Scan implements sql.Scanner for reading from NUMERIC column.
func (p *Price) Scan(value any) error {
	if value == nil {
		p.value = ""
		p.valid = false
		return nil
	}

	switch v := value.(type) {
	case string:
		p.value = v
		p.valid = true
	case []byte:
		p.value = string(v)
		p.valid = true
	case float64:
		p.value = strconv.FormatFloat(v, 'f', -1, 64)
		p.valid = true
	case int64:
		p.value = strconv.FormatInt(v, 10)
		p.valid = true
	default:
		return fmt.Errorf("cannot scan %T into Price", value)
	}

	if p.value == "" {
		p.valid = false
		return nil
	}
	rat, ok := new(big.Rat).SetString(p.value)
	if !ok {
		return fmt.Errorf("invalid price value scanned: %q", p.value)
	}
	p.value = normalizeToTwoDecimals(rat)
	p.valid = true
	return nil
}

// MarshalJSON encodes Price as a JSON number or null.
func (p Price) MarshalJSON() ([]byte, error) {
	if !p.valid {
		return []byte("null"), nil
	}
	return []byte(p.value), nil
}

// UnmarshalJSON decodes Price from a JSON number or null.
func (p *Price) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		p.value = ""
		p.valid = false
		return nil
	}
	s := string(data)
	if s == "null" {
		p.value = ""
		p.valid = false
		return nil
	}

	rat, ok := new(big.Rat).SetString(s)
	if !ok {
		return fmt.Errorf("invalid price JSON number: %s", s)
	}
	if rat.Sign() < 0 {
		return fmt.Errorf("price must be non-negative: %s", s)
	}

	p.value = normalizeToTwoDecimals(rat)
	p.valid = true
	return nil
}

// GORMDataType returns the GORM data type for Price.
func (Price) GORMDataType() string {
	return "numeric"
}

// GORMDBDataType returns the database column type for Price.
func (Price) GORMDBDataType(*gorm.DB) string {
	return "numeric(15,2)"
}

// normalizeToTwoDecimals normalizes a big.Rat to two decimal places string.
func normalizeToTwoDecimals(rat *big.Rat) string {
	cents := new(big.Int).Mul(rat.Num(), big.NewInt(100))
	cents.Quo(cents, rat.Denom())

	neg := false
	if cents.Sign() < 0 {
		neg = true
		cents.Neg(cents)
	}

	intPart := new(big.Int).Div(cents, big.NewInt(100))
	fracPart := new(big.Int).Mod(cents, big.NewInt(100))

	var sb strings.Builder
	if neg {
		sb.WriteByte('-')
	}
	sb.WriteString(intPart.String())
	sb.WriteByte('.')
	frac := fracPart.String()
	if len(frac) == 1 {
		sb.WriteString("0" + frac)
	} else {
		sb.WriteString(frac)
	}
	return sb.String()
}
