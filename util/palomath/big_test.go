package palomath_test

import (
	"math/big"
	"testing"

	"cosmossdk.io/math"
	"github.com/palomachain/paloma/util/palomath"
)

func TestBigetDiv(t *testing.T) {
	testCases := []struct {
		a, b     *big.Int
		expected string
	}{
		{big.NewInt(10), big.NewInt(2), "5.000000000000000000"},
		{big.NewInt(12345), big.NewInt(5678), "2.174181050000000000"},
		{big.NewInt(-12345), big.NewInt(5678), "-2.174181050000000000"},
	}

	for _, tc := range testCases {
		result := palomath.BigIntDiv(tc.a, tc.b)
		if result.String() != tc.expected {
			t.Errorf("BigetDiv(%v, %v) = %v, want %v", tc.a, tc.b, result, tc.expected)
		}
	}
}

func TestLegacyDecFromFloat64(t *testing.T) {
	tests := []struct {
		float64  float64
		expected math.LegacyDec
	}{
		{0.0, math.LegacyMustNewDecFromStr("0.0")},
		{1.0, math.LegacyMustNewDecFromStr("1.0")},
		{123.456, math.LegacyMustNewDecFromStr("123.456")},
		{-123.456, math.LegacyMustNewDecFromStr("-123.456")},
	}

	for _, tt := range tests {
		actual := palomath.LegacyDecFromFloat64(tt.float64)
		if !actual.Equal(tt.expected) {
			t.Errorf("LegacyDecFromFloat64(%f): expected %s, got %s", tt.float64, tt.expected, actual)
		}
	}
}
