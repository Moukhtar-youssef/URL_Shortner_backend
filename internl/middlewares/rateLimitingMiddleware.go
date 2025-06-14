package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

var (
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})
)

func AllowRequests(limit int, window time.Duration, banlimit int, banduration time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			key := "rate: " + ip
			bankey := "ban: " + ip
			isBanned, err := rdb.Exists(ctx, bankey).Result()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("Redis error: %w", err))
			}
			if isBanned == 1 {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"message": "ip has been bannd temporarly for too many requests",
				})
			}
			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "rate limiting error")
			}
			if count == 1 {
				rdb.Expire(ctx, key, window)
			}
			if count > int64(limit) {
				if count > int64(banlimit) {
					rdb.Set(ctx, bankey, "1", banduration)
					return c.JSON(http.StatusTooManyRequests, map[string]any{
						"Message": "Too many requests , IP temporarily banned.",
					})
				}

				reset, _ := rdb.TTL(ctx, key).Result()
				return c.JSON(http.StatusOK, map[string]any{
					"message":      "Rate limit exceeded",
					"retry_after":  reset.Seconds(),
					"limit":        limit,
					"current_hits": count,
				})
			}
			return next(c)
		}
	}
}
