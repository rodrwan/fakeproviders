package utils

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// NewID creates a new UUID.
func NewID() string {
	u2, err := uuid.NewRandom()
	if err != nil {
		fmt.Printf("Something went wrong: %s", err)
		return ""
	}
	return u2.String()
}

func RandomProcessTime(min, max int) time.Duration {
	rand.Seed(time.Now().UTC().UnixNano())
	return time.Duration(rand.Intn(max-min) + min)
}

func RandomStringNumber(n int) string {
	rand.Seed(time.Now().UnixNano())
	var numbers = []rune("0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = numbers[rand.Intn(len(numbers))]
	}
	return string(b)
}

func PickMonth() string {
	rand.Seed(time.Now().UnixNano())
	var months = []string{
		"01", "02", "03", "04", "05", "06", "07", "08", "09", "10", "11", "12",
	}
	return months[rand.Intn(len(months))]
}

func PickYear() string {
	rand.Seed(time.Now().UnixNano())
	var years = []string{
		"19", "20", "21", "22", "23",
	}
	return years[rand.Intn(len(years))]
}
