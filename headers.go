package ratelimit

import (
	"strconv"
	"time"
)

type (
	Headers    map[string]string
	HeaderKeys struct {
		XRateLimitRemaining  string
		XRateLimitLimit      string
		XRateLimitRetryAfter string
	}
)

var defaultHeaderKeys = HeaderKeys{
	XRateLimitRemaining:  "X-Ratelimit-Remaining",
	XRateLimitLimit:      "X-Ratelimit-Limit",
	XRateLimitRetryAfter: "X-Ratelimit-Retry-After",
}

func newHeaders(opts *Options, timestamp time.Time, curWindow int64, rate float64) Headers {
	headers := make(map[string]string)

	headers[headerKey(opts.Headers.XRateLimitLimit,
		defaultHeaderKeys.XRateLimitLimit)] = strconv.Itoa(opts.Rate)
	headers[headerKey(opts.Headers.XRateLimitRemaining,
		defaultHeaderKeys.XRateLimitRemaining)] = strconv.Itoa(opts.Rate - int(rate) - 1)

	xRateLimitRetryAfter := headerKey(opts.Headers.XRateLimitRetryAfter, defaultHeaderKeys.XRateLimitRetryAfter)
	if float64(opts.Rate) <= rate {
		sub := opts.Epoch.Add(time.Duration(curWindow+1) * opts.Window).Sub(timestamp).Milliseconds()
		headers[xRateLimitRetryAfter] = strconv.FormatInt(sub, 10)
	} else {
		headers[xRateLimitRetryAfter] = strconv.Itoa(0)
	}

	return headers
}

func headerKey(k, d string) string {
	if len(k) <= 0 {
		return d
	}

	return k
}
