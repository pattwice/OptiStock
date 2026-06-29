package alert

import "math/big"

func parseDecimal(s string) (*big.Rat, bool) {
	r := new(big.Rat)
	if _, ok := r.SetString(s); !ok {
		return nil, false
	}
	return r, true
}
