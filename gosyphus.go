// Package gosyphus implements jittered capped exponential backoff and provides
// functionality for retrying failed function calls.
//
// The top-level functions retry using the default wait time parameters.
// If you need to configure the wait time parameters, create a new [Gosyphus].
//
// When calling functions from this package, you can use [context] to set an overall
// deadline on the call or to cancel retries at any time.
//
// All functions in this package are safe for concurrent use by multiple goroutines.
//
// To learn more about jittered capped exponential backoff, see the article
// [Exponential Backoff And Jitter]. This package implements the “Full Jitter”
// algorithm mentioned in the article.
//
// [Exponential Backoff And Jitter]: https://aws.amazon.com/blogs/architecture/exponential-backoff-and-jitter/
package gosyphus

import (
	"context"
	"math/rand"
	"time"
)

// Gosyphus implements jittered capped exponential backoff.
type Gosyphus struct {
	initial time.Duration
	max     time.Duration
}

// New returns a new Gosyphus. The wait time between retries is a random value between
// 0 and d, where d starts at initial and doubles every retry, but is capped at max.
//
// If initial < 1, it will be set to 1. If max < initial, max will be set to initial.
func New(initial, max time.Duration) *Gosyphus {
	if initial < 1 {
		initial = 1
	}
	if max < initial {
		max = initial
	}
	return &Gosyphus{initial, max}
}

var defaultGosyphus = &Gosyphus{
	initial: 1 * time.Second,
	max:     30 * time.Second,
}

// Do calls f and retries until it succeeds. The wait time between retries is a random
// value between 0 and d, where d starts at 1 second and doubles every retry, but is
// capped at 30 seconds.
//
// If the provided context expires before f succeeds, Do returns the context's error.
// If the provided context has already expired when Do is called, Do will not call f.
func Do(ctx context.Context, f func() error) error {
	return defaultGosyphus.Do(ctx, f)
}

// Dos calls f and retries until it succeeds or returns an error for which shouldRetry
// returns false. The wait time between retries is a random value between 0 and d, where
// d starts at 1 second and doubles every retry, but is capped at 30 seconds.
//
// If shouldRetry returns false for an error returned by f, Dos returns that error.
// If the provided context expires before f succeeds, Dos returns the context's error.
// If the provided context has already expired when Dos is called, Dos will not call f.
func Dos(ctx context.Context, f func() error, shouldRetry func(error) bool) error {
	return defaultGosyphus.Dos(ctx, f, shouldRetry)
}

// Do calls f and retries until it succeeds.
//
// If the provided context expires before f succeeds, Do returns the context's error.
// If the provided context has already expired when Do is called, Do will not call f.
func (g *Gosyphus) Do(ctx context.Context, f func() error) error {
	alwaysRetry := func(error) bool { return true }
	return g.Dos(ctx, f, alwaysRetry)
}

// Dos calls f and retries until it succeeds or returns an error for which shouldRetry
// returns false.
//
// If shouldRetry returns false for an error returned by f, Dos returns that error.
// If the provided context expires before f succeeds, Dos returns the context's error.
// If the provided context has already expired when Dos is called, Dos will not call f.
func (g *Gosyphus) Dos(ctx context.Context, f func() error, shouldRetry func(error) bool) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if err := f(); err == nil {
		return nil
	} else if !shouldRetry(err) {
		return err
	}
	d := g.initial
	t := time.NewTimer(jitter(d))
	for {
		select {
		case <-ctx.Done():
			if !t.Stop() {
				<-t.C
			}
			return ctx.Err()
		case <-t.C:
			if err := f(); err == nil {
				return nil
			} else if !shouldRetry(err) {
				return err
			}
			d *= 2
			// d <= 0 checks for an overflow, but it should never happen in practice
			if d > g.max || d <= 0 {
				d = g.max
			}
			t.Reset(jitter(d))
		}
	}
}

// jitter returns a random value between 0 and d.
func jitter(d time.Duration) time.Duration {
	return time.Duration(rand.Int63n(int64(d)))
}
