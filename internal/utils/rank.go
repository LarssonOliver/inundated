package utils

import (
	"fmt"
	"strings"
)

// rankDigits are the digits of a rank, in ascending byte order, so ranks
// compare correctly as plain strings (and under Postgres's "C" collation).
const rankDigits = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// MaxRankLength is the rank length past which callers should respace a
// sibling list with [EvenRanks] rather than keep inserting between
// ever-closer neighbours.
const MaxRankLength = 32

// RankBetween returns a rank that sorts strictly between a and b. An empty a
// means "before everything" and an empty b means "after everything". Both
// must be valid ranks (made of rankDigits, not ending in '0') with a < b when
// both are set.
func RankBetween(a, b string) (string, error) {
	if !validRank(a) || !validRank(b) {
		return "", fmt.Errorf("RankBetween: invalid rank %q or %q", a, b)
	}
	if b != "" && a >= b {
		return "", fmt.Errorf("RankBetween: %q is not before %q", a, b)
	}
	return rankMidpoint(a, b, b != ""), nil
}

// rankMidpoint implements the midpoint step of fractional indexing: a string
// between a and b (or above a, when hasUpper is false), treating a missing
// digit in a as '0'.
func rankMidpoint(a, b string, hasUpper bool) string {
	if hasUpper {
		n := 0
		for n < len(b) && digitAt(a, n) == b[n] {
			n++
		}
		if n > 0 {
			return b[:n] + rankMidpoint(suffix(a, n), b[n:], true)
		}
	}

	low := 0
	if a != "" {
		low = strings.IndexByte(rankDigits, a[0])
	}
	high := len(rankDigits)
	if hasUpper {
		high = strings.IndexByte(rankDigits, b[0])
	}

	if high-low > 1 {
		return string(rankDigits[(low+high+1)/2])
	}
	if hasUpper && len(b) > 1 {
		return b[:1]
	}
	return string(rankDigits[low]) + rankMidpoint(suffix(a, 1), "", false)
}

// EvenRanks returns n ascending ranks spread evenly across the rank space,
// all of the same short length, for respacing a crowded sibling list.
func EvenRanks(n int) []string {
	if n <= 0 {
		return nil
	}
	width := 1
	space := len(rankDigits)
	for space <= n {
		width++
		space *= len(rankDigits)
	}

	ranks := make([]string, n)
	for i := range n {
		value := (i + 1) * space / (n + 1)
		digits := make([]byte, width)
		for d := width - 1; d >= 0; d-- {
			digits[d] = rankDigits[value%len(rankDigits)]
			value /= len(rankDigits)
		}
		ranks[i] = strings.TrimRight(string(digits), "0")
	}
	return ranks
}

func digitAt(s string, i int) byte {
	if i < len(s) {
		return s[i]
	}
	return '0'
}

func suffix(s string, n int) string {
	if n >= len(s) {
		return ""
	}
	return s[n:]
}

func validRank(s string) bool {
	if strings.HasSuffix(s, "0") {
		return false
	}
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(rankDigits, s[i]) == -1 {
			return false
		}
	}
	return true
}
