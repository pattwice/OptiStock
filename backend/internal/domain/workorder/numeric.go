package workorder

import (
	"math/big"
	"strings"
)

func parseRat(v string) (*big.Rat, bool) {
	return new(big.Rat).SetString(strings.TrimSpace(v))
}

func ratString(r *big.Rat) string {
	if r == nil {
		return "0"
	}
	return r.FloatString(6)
}

func addRat(a, b string) (string, bool) {
	ra, ok := parseRat(a)
	if !ok {
		return "", false
	}
	rb, ok := parseRat(b)
	if !ok {
		return "", false
	}
	return ratString(new(big.Rat).Add(ra, rb)), true
}

func subRat(a, b string) (string, bool) {
	ra, ok := parseRat(a)
	if !ok {
		return "", false
	}
	rb, ok := parseRat(b)
	if !ok {
		return "", false
	}
	return ratString(new(big.Rat).Sub(ra, rb)), true
}

func cmpRat(a, b string) (int, bool) {
	ra, ok := parseRat(a)
	if !ok {
		return 0, false
	}
	rb, ok := parseRat(b)
	if !ok {
		return 0, false
	}
	return ra.Cmp(rb), true
}

func mulRat(a, b string) (string, bool) {
	ra, ok := parseRat(a)
	if !ok {
		return "", false
	}
	rb, ok := parseRat(b)
	if !ok {
		return "", false
	}
	return ratString(new(big.Rat).Mul(ra, rb)), true
}

func negateQty(v string) (string, bool) {
	r, ok := parseRat(v)
	if !ok || r.Sign() <= 0 {
		return "", false
	}
	r.Neg(r)
	return ratString(r), true
}

func isPositive(v string) bool {
	r, ok := parseRat(v)
	return ok && r.Sign() > 0
}

func isZeroOrEmpty(v string) bool {
	if strings.TrimSpace(v) == "" {
		return true
	}
	r, ok := parseRat(v)
	return ok && r.Sign() == 0
}
