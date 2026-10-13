package permission

import (
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"
)

const (
	Allow = 1
	Deny  = 0
)

type BigPerm struct {
	value *big.Int
}

func NewFromBigInt(value *big.Int) *BigPerm {
	return &BigPerm{
		value: value,
	}
}

func NewFromPerm(p *BigPerm) *BigPerm {
	return p.Clone()
}

func NewFromStr(value string) *BigPerm {
	bit := new(big.Int)
	if value == "" {
		return NewFromBigInt(bit)
	}
	bit.SetString(value, 10)
	return NewFromBigInt(bit)
}

func NewFromUint64(value uint64) *BigPerm {
	bit := new(big.Int)
	bit.SetUint64(value)
	return NewFromBigInt(bit)
}

func New() *BigPerm {
	return NewFromBigInt(big.NewInt(0))
}

func (p *BigPerm) Grant(bitIndex int) {
	p.value.SetBit(p.value, bitIndex, Allow)
}

func (p *BigPerm) Revoke(bitIndex int) {
	p.value.SetBit(p.value, bitIndex, Deny)
}

func (p *BigPerm) Union(other *BigPerm) {
	p.value.Or(p.value, other.value)
}

func (p *BigPerm) UnionStr(other string) {
	p.Union(NewFromStr(other))
}

func (p *BigPerm) Has(bitIndex int) bool {
	return p.value.Bit(bitIndex) == Allow
}

func (p *BigPerm) HasAll(bitIndices ...int) bool {
	for _, index := range bitIndices {
		if !p.Has(index) {
			return false
		}
	}
	return true
}

func (p *BigPerm) HasAllPerms(perms *BigPerm) bool {
	isAllowed := p.value.
		And(p.value, perms.value).
		Cmp(perms.value) == 0
	return isAllowed
}

func (p *BigPerm) HasOneOf(bitIndices ...int) bool {
	return slices.ContainsFunc(bitIndices, p.Has)
}

func (p *BigPerm) HasOneOfPerms(perms *BigPerm) bool {
	isAllowed := p.value.
		And(p.value, perms.value).
		Sign() == 1
	return isAllowed
}

func (p *BigPerm) Cannot(bitIndex int) bool {
	return p.value.Bit(bitIndex) == Deny
}

func (p *BigPerm) Value() *big.Int {
	return p.value
}

func (p *BigPerm) String() string {
	return p.value.String()
}

func (p *BigPerm) Uint64() uint64 {
	return p.value.Uint64()
}

func (p *BigPerm) GetSetBitIndices() []int {
	return getSetBitIndices(p.value)
}

func (p *BigPerm) Clone() *BigPerm {
	return NewFromBigInt(new(big.Int).Set(p.value))
}

func (p BigPerm) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, `"%s"`, p.value.String()), nil
}

func (p *BigPerm) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	bit := new(big.Int)
	_, ok := bit.SetString(s, 10)
	if !ok {
		return errors.New("invalid bitmask string")
	}
	p.value = bit
	return nil
}
