package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func RandomInt(min, max int) int {
	return min + rand.Intn(max-min+1)
}

func RandomDuration(min, max int) time.Duration {
	return time.Duration(RandomInt(min, max)) * time.Second
}

func RandomChoice(list []string) string {
	return list[rand.Intn(len(list))]
}

func RandomPerm(n int) []int {
	return rand.Perm(n)
}

func RandomJitter(base time.Duration, jitterPercent int) time.Duration {
	jitter := float64(base) * (float64(RandomInt(0, jitterPercent)) / 100.0)
	if rand.Intn(2) == 0 {
		return base + time.Duration(jitter)
	}
	return base - time.Duration(jitter)
}
