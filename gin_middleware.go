package ratelimit

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GinOptions struct {
	*Options
	Key func(*gin.Context) (string, error)
}

func GinAllow(ginOpts *GinOptions, r RateLimiter) gin.HandlerFunc {
	if ginOpts.Epoch.Nanosecond() <= 0 {
		ginOpts.Epoch = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	}

	return func(ctx *gin.Context) {
		key, err := ginOpts.Key(ctx)
		if err != nil {
			ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}

		statusCode, err := r.Allow(ctx.Request.Context(), ginOpts.Options, key)
		if err != nil || statusCode != http.StatusOK {
			ctx.AbortWithError(statusCode, err)
			return
		}

		ctx.Next()
	}
}
