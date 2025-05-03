package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// GenerateValidCardNumber creates a 16-digit number passing the Luhn algorithm.
func GenerateValidCardNumber() string {
	rand.Seed(time.Now().UnixNano())
	digits := make([]int, 15)

	// Начальные 6 цифр — BIN (например, 400000)
	copy(digits, []int{4, 0, 0, 0, 0, 0})
	for i := 6; i < 15; i++ {
		digits[i] = rand.Intn(10)
	}

	// Вычисляем контрольную цифру (16-я)
	sum := 0
	for i := 0; i < 15; i++ {
		d := digits[i]
		if i%2 == 0 {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
	}
	checkDigit := (10 - (sum % 10)) % 10

	digits = append(digits, checkDigit)
	result := ""
	for _, d := range digits {
		result += strconv.Itoa(d)
	}
	return result
}
