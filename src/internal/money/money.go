package money

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ParseCents converts a decimal money string into integer cents.
func ParseCents(value string) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("amount is empty")
	}

	sign := int64(1)
	if strings.HasPrefix(value, "-") {
		sign = -1
		value = strings.TrimPrefix(value, "-")
	}

	parts := strings.Split(value, ".")
	if len(parts) > 2 {
		return 0, fmt.Errorf("amount has too many decimal separators")
	}

	wholePart := parts[0]
	if wholePart == "" {
		wholePart = "0"
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("amount is invalid")
	}

	fraction := "00"
	if len(parts) == 2 {
		if len(parts[1]) > 2 {
			return 0, fmt.Errorf("amount has more than 2 decimal places")
		}

		fraction = parts[1]
		if len(fraction) == 1 {
			fraction += "0"
		}
		if fraction == "" {
			fraction = "00"
		}
	}

	cents, err := strconv.ParseInt(fraction, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("amount is invalid")
	}

	return sign * ((whole * 100) + cents), nil
}

// ToCents rounds a float amount into integer cents.
func ToCents(value float64) int64 {
	return int64(math.Round(value * 100))
}

// FromCents converts integer cents to a float amount.
func FromCents(value int64) float64 {
	return float64(value) / 100
}

// Normalize rounds a float amount to 2 decimal places.
func Normalize(value float64) float64 {
	return FromCents(ToCents(value))
}

// ParseAmount converts a decimal string into a normalized float amount.
func ParseAmount(value string) (float64, error) {
	cents, err := ParseCents(value)
	if err != nil {
		return 0, err
	}

	return FromCents(cents), nil
}

// SplitEvenly rounds an even split to the nearest cent.
func SplitEvenly(totalCents int64, parts int) int64 {
	if parts <= 0 {
		return 0
	}

	divisor := int64(parts)
	quotient := totalCents / divisor
	remainder := totalCents % divisor

	if remainder*2 >= divisor {
		quotient++
	}

	return quotient
}
