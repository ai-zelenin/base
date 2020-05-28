package util

import (
	mathRand "math/rand"
	"time"
)

var MathRandSeedSource = mathRand.NewSource(time.Now().UnixNano())

const factor = 1000000000.0
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// RandomString create string with n random chars
func RandomString(n int) string {
	b := make([]byte, n)
	for i, cache, remain := n-1, MathRandSeedSource.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = MathRandSeedSource.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// RandomRangeFloat64 create random float in given range
func RandomRangeFloat64(min, max float64) float64 {
	minInt := int64(min * factor)
	maxInt := int64(max * factor)
	random := mathRand.Int63n(maxInt - minInt)
	return float64(random+minInt) / factor
}

// WeightedRandom return random key from defined map but probability of each key depends of value
func WeightedRandom(pm map[int64]int64) int64 {
	var idLine = make([]int64, 0)
	for id, weight := range pm {
		for i := 0; int64(i) < weight; i++ {
			idLine = append(idLine, id)
		}
	}
	mathRand.Shuffle(len(idLine), func(i, j int) {
		idLine[i], idLine[j] = idLine[j], idLine[i]
	})
	index := mathRand.Intn(len(idLine))
	result := idLine[index]
	return result
}
