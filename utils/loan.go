package utils

import "math"

// Рассчитывает ежемесячный аннуитетный платёж
func CalculateAnnuityPayment(principal float64, annualRate float64, months int) float64 {
	monthlyRate := annualRate / 12.0 / 100.0
	payment := principal * (monthlyRate * math.Pow(1+monthlyRate, float64(months))) /
		(math.Pow(1+monthlyRate, float64(months)) - 1)
	return math.Round(payment*100) / 100 // округление до копеек
}
