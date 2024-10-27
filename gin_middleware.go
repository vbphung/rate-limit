package ratelimit

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type GinOptions struct {
	*Options
	Key func(*gin.Context) (string, error)
}

func GinAllow(ginOpts *GinOptions, r RateLimiter) gin.HandlerFunc {
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
