package auditlog

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const MaxLookback = 90 * 24 * time.Hour

func ParseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 24 * time.Hour, nil
	}

	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(s[:len(s)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		d := time.Duration(days) * 24 * time.Hour
		if d > MaxLookback {
			return MaxLookback, nil
		}
		return d, nil
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}
	if d > MaxLookback {
		return MaxLookback, nil
	}
	return d, nil
}
