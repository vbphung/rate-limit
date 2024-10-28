package ratelimit

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type req struct {
	t  time.Time
	ok bool
}

func TestRate(t *testing.T) {
	store := NewMemStore()
	r := New(store)
	opts := &Options{
		Rate:   1000,
		Window: time.Second * 3,
		Epoch:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	bytes := make([]byte, 12)

	_, err := rand.Read(bytes)
	require.NoError(t, err)

	var (
		n, m  = 100, 3
		reqs  = make([][]req, n)
		key   = base64.RawURLEncoding.EncodeToString(bytes)
		delay = time.Millisecond * 100
		wg    sync.WaitGroup
	)

	fmt.Println(key)

	sendReq := func(ctx context.Context, c int) {
		go func() {
			t := time.Now()
			_, statusCode, err := r.Allow(ctx, opts, key)
			reqs[c] = append(reqs[c], req{t, err == nil && statusCode == http.StatusOK})
		}()

		time.Sleep(delay)
	}

	for c := range n {
		time.Sleep(time.Millisecond * 1)
		wg.Add(1)

		go func(c int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), opts.Window*time.Duration(m))
			defer cancel()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					sendReq(ctx, c)
				}
			}
		}(c)
	}

	wg.Wait()

	var res []req
	for _, r := range reqs {
		res = append(res, r...)
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].t.Before(res[j].t)
	})

	var oks [][]int
	for i := 0; i < len(res); i++ {
		if !res[i].ok {
			continue
		}

		v := []int{i}
		for i < len(res) && res[i].ok {
			i++
		}

		if i-1 == v[0] {
			oks = append(oks, v)
		} else {
			oks = append(oks, append(v, i-1))
		}
	}

	bytes, err = json.Marshal(oks[:32])
	require.NoError(t, err)

	fmt.Println(string(bytes))
}
