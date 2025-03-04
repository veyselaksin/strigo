package duration

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

// Period represents predefined time periods for rate limiting
// Using string type for better readability and JSON serialization
type Period string

// Predefined periods for rate limiting
// Using *LY suffix for consistency and clarity
const (
	SECONDLY Period = "SECONDLY" // Rate limit per second
	MINUTELY Period = "MINUTELY" // Rate limit per minute
	HOURLY   Period = "HOURLY"   // Rate limit per hour
	DAILY    Period = "DAILY"    // Rate limit per day
	WEEKLY   Period = "WEEKLY"   // Rate limit per week
	MONTHLY  Period = "MONTHLY"  // Rate limit per month (30 days)
	YEARLY   Period = "YEARLY"   // Rate limit per year (365 days)
)

// RateLimit represents a rate limit duration like "5/MINUTE", "1/HOUR", etc.
type RateLimit struct {
	Count  int64
	Period Period
}

var durationRegex = regexp.MustCompile(`^(\d+)/([A-Z]+)$`)

// ParseRateLimit parses a rate limit string like "5/MINUTE" into a RateLimit struct
func ParseRateLimit(s string) (*RateLimit, error) {
	matches := durationRegex.FindStringSubmatch(s)
	if matches == nil {
		return nil, fmt.Errorf("invalid rate limit format: %s (expected format: COUNT/PERIOD)", s)
	}

	count, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid count: %s", matches[1])
	}

	period := Period(matches[2])
	if !IsValidPeriod(period) {
		return nil, fmt.Errorf("invalid period: %s (expected: SECONDLY, MINUTELY, HOURLY, DAILY, WEEKLY, MONTHLY)", period)
	}

	return &RateLimit{
		Count:  count,
		Period: period,
	}, nil
}

// ToDuration converts the period to a time.Duration
// This is used internally by the rate limiter to set key expiration
func (p Period) ToDuration() time.Duration {
	switch p {
	case SECONDLY:
		return time.Second
	case MINUTELY:
		return time.Minute
	case HOURLY:
		return time.Hour
	case DAILY:
		return 24 * time.Hour
	case WEEKLY:
		return 7 * 24 * time.Hour
	case MONTHLY:
		return 30 * 24 * time.Hour // Approximate month length
	case YEARLY:
		return 365 * 24 * time.Hour // Approximate year length
	default:
		return time.Minute // Safe default
	}
}

func IsValidPeriod(p Period) bool {
	switch p {
	case SECONDLY, MINUTELY, HOURLY, DAILY, WEEKLY, MONTHLY:
		return true
	default:
		return false
	}
}

// String returns the string representation of the rate limit
func (r *RateLimit) String() string {
	return fmt.Sprintf("%d/%s", r.Count, r.Period)
}
