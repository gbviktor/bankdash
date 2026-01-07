package util

import (
	"fmt"
	"strconv"
	"strings"
)

func ParseAmountCents(s, decimalMode, thousandsSep string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty amount")
	}

	neg := false
	if strings.HasPrefix(s, "-") {
		neg = true
		s = strings.TrimPrefix(s, "-")
	}
	if strings.HasPrefix(s, "+") {
		s = strings.TrimPrefix(s, "+")
	}
	s = strings.TrimSpace(s)

	// remove thousands separators
	if thousandsSep != "" {
		s = strings.ReplaceAll(s, thousandsSep, "")
	}

	// decimal comma for DE
	if decimalMode == "de" {
		s = strings.ReplaceAll(s, ",", ".")
	}

	// now s is like "28.95" or "28" or "28.9"
	wholePart := s
	fracPart := ""
	if dot := strings.IndexByte(s, '.'); dot >= 0 {
		wholePart = s[:dot]
		fracPart = s[dot+1:]
	}

	if wholePart == "" {
		wholePart = "0"
	}

	whole, err := strconv.ParseInt(wholePart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid whole part: %q", wholePart)
	}

	// normalize fractional to 2 digits
	if fracPart == "" {
		fracPart = "00"
	} else if len(fracPart) == 1 {
		fracPart = fracPart + "0"
	} else if len(fracPart) > 2 {
		fracPart = fracPart[:2]
	}

	frac, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid fractional part: %q", fracPart)
	}

	cents := whole*100 + frac
	if neg {
		cents = -cents
	}
	return cents, nil
}
