package strigo

import (
	"context"
	"fmt"
	"math"
	"time"
)

// Strategy-specific data structures

// TokenBucketData represents the state of a token bucket
type TokenBucketData struct {
	Tokens     float64   `json:"tokens"`
	LastRefill time.Time `json:"last_refill"`
	Capacity   int64     `json:"capacity"`
	RefillRate float64   `json:"refill_rate"`
}

// LeakyBucketData represents the state of a leaky bucket
type LeakyBucketData struct {
	Queue     []QueuedRequest `json:"queue"`
	LastDrain time.Time       `json:"last_drain"`
	DrainRate float64         `json:"drain_rate"`
}

// QueuedRequest represents a request in the leaky bucket queue
type QueuedRequest struct {
	Timestamp time.Time `json:"timestamp"`
	Points    int64     `json:"points"`
}

// SlidingWindowData represents the state of a sliding window
type SlidingWindowData struct {
	Requests []time.Time `json:"requests"`
}

// FixedWindowData represents the state of a fixed window (legacy for compatibility)
type FixedWindowData struct {
	Count      int64     `json:"count"`
	WindowStart time.Time `json:"window_start"`
}

// Strategy-specific implementations

// consumeTokenBucket implements the classic token bucket algorithm
func (rl *RateLimiter) consumeTokenBucket(ctx context.Context, key string, points int64) (*Result, error) {
	now := time.Now()
	storageKey := rl.buildKey(key)
	dataKey := fmt.Sprintf("%s:tb", storageKey)
	
	// Get current bucket state
	var data TokenBucketData
	err := rl.storage.GetJSON(ctx, dataKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get token bucket data: %w", err)
	}
	
	// Initialize if first time
	if data.LastRefill.IsZero() {
		data.Capacity = rl.opts.Points
		data.RefillRate = float64(rl.opts.Points) / rl.opts.GetDuration().Seconds()
		data.Tokens = float64(rl.opts.Points) // Start with full bucket
		data.LastRefill = now
	}
	
	// Calculate tokens to add based on elapsed time
	elapsed := now.Sub(data.LastRefill).Seconds()
	tokensToAdd := elapsed * data.RefillRate
	data.Tokens = math.Min(float64(data.Capacity), data.Tokens+tokensToAdd)
	data.LastRefill = now
	
	// Check if enough tokens available
	if data.Tokens >= float64(points) {
		data.Tokens -= float64(points)
		
		// Save updated state
		err = rl.storage.SetJSON(ctx, dataKey, data, rl.opts.GetDuration()*2)
		if err != nil {
			return nil, fmt.Errorf("failed to save token bucket data: %w", err)
		}
		
		return &Result{
			MsBeforeNext:      0,
			RemainingPoints:   int64(data.Tokens),
			ConsumedPoints:    points,
			IsFirstInDuration: elapsed > rl.opts.GetDuration().Seconds(),
			TotalHits:         rl.opts.Points,
			Allowed:           true,
		}, nil
	}
	
	// Calculate time until enough tokens are available
	tokensNeeded := float64(points) - data.Tokens
	msBeforeNext := int64((tokensNeeded / data.RefillRate) * 1000)
	
	return &Result{
		MsBeforeNext:      msBeforeNext,
		RemainingPoints:   int64(data.Tokens),
		ConsumedPoints:    0,
		IsFirstInDuration: false,
		TotalHits:         rl.opts.Points,
		Allowed:           false,
	}, nil
}

// consumeLeakyBucket implements the leaky bucket algorithm
func (rl *RateLimiter) consumeLeakyBucket(ctx context.Context, key string, points int64) (*Result, error) {
	now := time.Now()
	storageKey := rl.buildKey(key)
	dataKey := fmt.Sprintf("%s:lb", storageKey)
	
	// Get current bucket state
	var data LeakyBucketData
	err := rl.storage.GetJSON(ctx, dataKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaky bucket data: %w", err)
	}
	
	// Initialize if first time
	if data.LastDrain.IsZero() {
		data.DrainRate = float64(rl.opts.Points) / rl.opts.GetDuration().Seconds()
		data.LastDrain = now
		data.Queue = make([]QueuedRequest, 0)
	}
	
	// Drain bucket based on elapsed time
	elapsed := now.Sub(data.LastDrain).Seconds()
	requestsToDrain := int64(elapsed * data.DrainRate)
	data.Queue = rl.drainRequests(data.Queue, requestsToDrain)
	data.LastDrain = now
	
	// Calculate current queue size in points
	currentPoints := int64(0)
	for _, req := range data.Queue {
		currentPoints += req.Points
	}
	
	// Check if bucket has capacity
	if currentPoints+points <= rl.opts.Points {
		// Add to queue
		data.Queue = append(data.Queue, QueuedRequest{
			Timestamp: now,
			Points:    points,
		})
		
		// Save updated state
		err = rl.storage.SetJSON(ctx, dataKey, data, rl.opts.GetDuration()*2)
		if err != nil {
			return nil, fmt.Errorf("failed to save leaky bucket data: %w", err)
		}
		
		return &Result{
			MsBeforeNext:      0,
			RemainingPoints:   rl.opts.Points - (currentPoints + points),
			ConsumedPoints:    currentPoints + points,
			IsFirstInDuration: len(data.Queue) == 1,
			TotalHits:         rl.opts.Points,
			Allowed:           true,
		}, nil
	}
	
	// Calculate delay based on drain rate
	pointsOverflow := (currentPoints + points) - rl.opts.Points
	msBeforeNext := int64((float64(pointsOverflow) / data.DrainRate) * 1000)
	
	return &Result{
		MsBeforeNext:      msBeforeNext,
		RemainingPoints:   rl.opts.Points - currentPoints,
		ConsumedPoints:    currentPoints,
		IsFirstInDuration: false,
		TotalHits:         rl.opts.Points,
		Allowed:           false,
	}, nil
}

// consumeSlidingWindow implements the sliding window algorithm
func (rl *RateLimiter) consumeSlidingWindow(ctx context.Context, key string, points int64) (*Result, error) {
	now := time.Now()
	storageKey := rl.buildKey(key)
	dataKey := fmt.Sprintf("%s:sw", storageKey)
	windowStart := now.Add(-rl.opts.GetDuration())
	
	// Get current window state
	var data SlidingWindowData
	err := rl.storage.GetJSON(ctx, dataKey, &data)
	if err != nil {
		return nil, fmt.Errorf("failed to get sliding window data: %w", err)
	}
	
	// Initialize if first time
	if data.Requests == nil {
		data.Requests = make([]time.Time, 0)
	}
	
	// Remove old requests outside window
	data.Requests = rl.removeOldRequests(data.Requests, windowStart)
	
	// Check if adding new requests would exceed limit
	if int64(len(data.Requests))+points <= rl.opts.Points {
		// Add new request timestamps
		for i := int64(0); i < points; i++ {
			data.Requests = append(data.Requests, now)
		}
		
		// Save updated state
		err = rl.storage.SetJSON(ctx, dataKey, data, rl.opts.GetDuration()*2)
		if err != nil {
			return nil, fmt.Errorf("failed to save sliding window data: %w", err)
		}
		
		return &Result{
			MsBeforeNext:      0,
			RemainingPoints:   rl.opts.Points - int64(len(data.Requests)),
			ConsumedPoints:    int64(len(data.Requests)),
			IsFirstInDuration: len(data.Requests) == int(points),
			TotalHits:         rl.opts.Points,
			Allowed:           true,
		}, nil
	}
	
	// Calculate time until oldest request expires
	if len(data.Requests) > 0 {
		oldestRequest := data.Requests[0]
		msBeforeNext := oldestRequest.Add(rl.opts.GetDuration()).Sub(now).Milliseconds()
		if msBeforeNext < 0 {
			msBeforeNext = 0
		}
		
		return &Result{
			MsBeforeNext:      msBeforeNext,
			RemainingPoints:   rl.opts.Points - int64(len(data.Requests)),
			ConsumedPoints:    int64(len(data.Requests)),
			IsFirstInDuration: false,
			TotalHits:         rl.opts.Points,
			Allowed:           false,
		}, nil
	}
	
	return &Result{
		MsBeforeNext:      0,
		RemainingPoints:   rl.opts.Points,
		ConsumedPoints:    0,
		IsFirstInDuration: true,
		TotalHits:         rl.opts.Points,
		Allowed:           false,
	}, nil
}

// consumeFixedWindow implements the fixed window algorithm (existing implementation)
func (rl *RateLimiter) consumeFixedWindow(ctx context.Context, key string, points int64) (*Result, error) {
	storageKey := rl.buildKey(key)
	
	// Get current window information
	windowStart := rl.getWindowStartFixed()
	windowKey := fmt.Sprintf("%s:%d", storageKey, windowStart.Unix())
	
	// Get current count from storage
	currentCount, err := rl.storage.Get(ctx, windowKey)
	if err != nil {
		return nil, fmt.Errorf("failed to get current count: %w", err)
	}
	
	// Check if this is the first request in the window
	isFirstInDuration := currentCount == 0
	
	// Calculate if the request should be allowed
	newCount := currentCount + points
	allowed := newCount <= rl.opts.Points
	
	// Calculate remaining points
	remainingPoints := rl.opts.Points - currentCount
	if remainingPoints < 0 {
		remainingPoints = 0
	}
	
	// Calculate time until next window
	nextWindow := windowStart.Add(rl.opts.GetDuration())
	msBeforeNext := time.Until(nextWindow).Milliseconds()
	
	// If allowed, increment the counter
	consumedPoints := currentCount
	if allowed {
		_, err = rl.storage.Increment(ctx, windowKey, points, rl.opts.GetDuration())
		if err != nil {
			return nil, fmt.Errorf("failed to increment counter: %w", err)
		}
		consumedPoints = newCount
		remainingPoints = rl.opts.Points - newCount
		if remainingPoints < 0 {
			remainingPoints = 0
		}
	}
	
	result := &Result{
		MsBeforeNext:      msBeforeNext,
		RemainingPoints:   remainingPoints,
		ConsumedPoints:    consumedPoints,
		IsFirstInDuration: isFirstInDuration,
		TotalHits:         rl.opts.Points,
		Allowed:           allowed,
	}
	
	return result, nil
}

// Helper functions

// drainRequests removes the specified number of requests from the queue
func (rl *RateLimiter) drainRequests(queue []QueuedRequest, requestsToDrain int64) []QueuedRequest {
	pointsDrained := int64(0)
	drainIndex := 0
	
	for i, req := range queue {
		if pointsDrained >= requestsToDrain {
			break
		}
		pointsDrained += req.Points
		drainIndex = i + 1
	}
	
	if drainIndex >= len(queue) {
		return make([]QueuedRequest, 0)
	}
	
	return queue[drainIndex:]
}

// removeOldRequests removes requests that are outside the sliding window
func (rl *RateLimiter) removeOldRequests(requests []time.Time, windowStart time.Time) []time.Time {
	validRequests := make([]time.Time, 0)
	
	for _, req := range requests {
		if req.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}
	
	return validRequests
}

// getWindowStartFixed returns the start time for fixed window strategy
func (rl *RateLimiter) getWindowStartFixed() time.Time {
	now := time.Now()
	duration := rl.opts.GetDuration()
	return now.Truncate(duration)
} 