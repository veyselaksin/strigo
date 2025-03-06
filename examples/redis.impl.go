package examples

import (
	"github.com/veyselaksin/strigo"
)

func NewRedisLimiter(cfg strigo.LimiterConfig) (strigo.Limiter, error) {
	return strigo.NewLimiter(cfg)
}
