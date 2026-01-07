package util

import (
	"fmt"
	"strings"
	"time"
)

func ParseDate(s string, formats []string, loc *time.Location) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date")
	}
	if loc == nil {
		loc = time.UTC
	}
	for _, f := range formats {
		if t, err := time.ParseInLocation(f, s, loc); err == nil {
			// normalize to midnight local time for deterministic day grouping
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported date format %q (tried %v)", s, formats)
}
