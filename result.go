package strigo

import (
	"strconv"
	"time"
)

// Result represents the result of a rate limit operation
// Similar to RateLimiterRes from rate-limiter-flexible package
type Result struct {
	// Number of milliseconds before next action can be done
	MsBeforeNext int64 `json:"msBeforeNext"`
	
	// Number of remaining points in current duration
	RemainingPoints int64 `json:"remainingPoints"`
	
	// Number of consumed points in current duration
	ConsumedPoints int64 `json:"consumedPoints"`
	
	// Whether the action is first in current duration
	IsFirstInDuration bool `json:"isFirstInDuration"`
	
	// Total points allowed in the duration
	TotalHits int64 `json:"totalHits"`
	
	// Whether the request was allowed
	Allowed bool `json:"allowed"`
}

// Headers returns HTTP headers that can be set in HTTP responses
// following common rate limiting header conventions
func (r *Result) Headers() map[string]string {
	headers := make(map[string]string)
	
	headers["X-RateLimit-Limit"] = toStr(r.TotalHits)
	headers["X-RateLimit-Remaining"] = toStr(r.RemainingPoints)
	headers["X-RateLimit-Reset"] = toStr(time.Now().Add(time.Duration(r.MsBeforeNext) * time.Millisecond).Unix())
	
	if !r.Allowed {
		headers["Retry-After"] = toStr(r.MsBeforeNext / 1000)
	}
	
	return headers
}

// Helper function to convert int64 to string
func toStr(i int64) string {
	return strconv.FormatInt(i, 10)
} 