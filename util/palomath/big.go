package palomath

import (
	"math/big"

	"cosmossdk.io/math"
)

// cDivisionPrecision specifies the precision of the quotient of
// eukledian divisions used during performance scoring.
const cDivisionPrecision int = 8

// BigIntDiv
// Quo returns the rounded quotient a/b.
// Precision is set to cDivisionPrecision.
// Panics if both operands are zero or infinities.
func BigIntDiv(a, b *big.Int) math.LegacyDec {
	i := big.NewFloat(0).SetInt(a)
	j := big.NewFloat(0).SetInt(b)

	result := big.NewFloat(0).Quo(i, j)
	return bigFloatToLegacyDec(result)
}

func LegacyDecFromFloat64(f float64) math.LegacyDec {
	return bigFloatToLegacyDec(big.NewFloat(f))
}

func bigFloatToLegacyDec(f *big.Float) math.LegacyDec {
	return math.LegacyMustNewDecFromStr(f.Text('f', cDivisionPrecision))
}
