package ratelimit

import (
	"strconv"
	"time"
)

type (
	header  string
	Headers map[header]string
)

const (
	XRateLimitRemaining  header = "X-Ratelimit-Remaining"
	XRateLimitLimit      header = "X-Ratelimit-Limit"
	XRateLimitRetryAfter header = "X-Ratelimit-Retry-After"
)

func newHeaders(opts *Options, timestamp time.Time, curWindow int64, rate float64) Headers {
	headers := make(map[header]string)

	headers[XRateLimitLimit] = strconv.Itoa(opts.Rate)
	headers[XRateLimitRemaining] = strconv.Itoa(opts.Rate - int(rate) - 1)

	if float64(opts.Rate) <= rate {
		sub := opts.Epoch.Add(time.Duration(curWindow+1) * opts.Window).Sub(timestamp).Milliseconds()
		headers[XRateLimitRetryAfter] = strconv.FormatInt(sub, 10)
	} else {
		headers[XRateLimitRetryAfter] = strconv.Itoa(0)
	}

	return headers
}
