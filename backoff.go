package mongoproxy

import (
	"math/rand"
	"time"
)

/*
Interface BackoffPolicy can be used to compute the next backoff value when
retrying requests to a server or other contended resource. For instance:

	for {
		if ok := sendRequest(); ok {
			break
		}

		next, ok := policy.Next(); ok {
			time.Sleep(next)
		} else {
			break
		}
	}
*/
type BackoffPolicy interface {
	Next() (next time.Duration, ok bool)
}

// A backoff policy that always returns the same value for Next()
type ConstantBackoffPolicy struct {
	i, maxRetries int
	interval      time.Duration
}

func NewConstantBackoffPolicy(maxRetries int, interval time.Duration) BackoffPolicy {
	return &ConstantBackoffPolicy{maxRetries: maxRetries, interval: interval}
}

func (p *ConstantBackoffPolicy) Next() (next time.Duration, ok bool) {
	if p.i < p.maxRetries {
		ok = true
		next = p.interval
		p.i++
	}

	return
}

// A backoff policy that returns a value in 2^i * interval + rand[0, interval)
// for Next(), where i is the index of the retry
type ExpBackoffPolicy struct {
	i, maxRetries int
	interval      time.Duration
}

func NewExpBackoffPolicy(maxRetries int, interval time.Duration) BackoffPolicy {
	return &ExpBackoffPolicy{maxRetries: maxRetries, interval: interval}
}

func (p *ExpBackoffPolicy) Next() (next time.Duration, ok bool) {
	if p.i < p.maxRetries {
		ok = true
		cur := int64(p.interval)
		next = time.Duration((1<<uint(p.i))*cur + rand.Int63n(cur))
		p.i++
	}

	return
}

// A helper which sleeps with the value returned by policy.Next() as long as the
// block provided returns true.
func CallWithBackoff(block func() bool, policy BackoffPolicy) {
	for {
		if !block() {
			break
		}

		if next, ok := policy.Next(); ok {
			time.Sleep(next)
		} else {
			break
		}
	}
}
