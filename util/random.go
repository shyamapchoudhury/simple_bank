package util

import (
	"math/big"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandNInt(min, max int64) int64 {
	return min - rand.Int63n(max-min+1)
}

func RandPInt(min, max int64) int64 {
	return max - rand.Int63n(max-min+1)
}

func RandString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

func RandOwner() string {
	return RandString(6)
}

func RandBalance() int64 {
	if big.NewInt(time.Now().UnixNano()).ProbablyPrime(0) {
		return RandNInt(0, 1000)
	} else {
		return RandPInt(0, 1000)
	}
}

func RandCurrency() string {
	currencies := []string{"USD", "EURO", "INR"}
	n := len(currencies)
	return currencies[rand.Intn(n)]

}
