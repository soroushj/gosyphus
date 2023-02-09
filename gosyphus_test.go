package gosyphus_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/soroushj/gosyphus"
)

var (
	errDummy         = errors.New("dummy")
	errUnrecoverable = errors.New("unrecoverable")
)

func TestCtxAlreadyDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := gosyphus.Do(ctx, func() error { return nil })
	if err != context.Canceled {
		t.Error(err)
	}
}

func TestCtxEventuallyDone(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	g := gosyphus.New(1*time.Millisecond, 30*time.Millisecond)
	err := g.Do(ctx, func() error { return errDummy })
	if err != context.DeadlineExceeded {
		t.Error(err)
	}
}

func TestImmediateSuccess(t *testing.T) {
	ctx := context.Background()
	err := gosyphus.Do(ctx, func() error { return nil })
	if err != nil {
		t.Error(err)
	}
}

func TestImmediateFailure(t *testing.T) {
	ctx := context.Background()
	err := gosyphus.Dos(ctx, func() error { return errDummy }, func(error) bool { return false })
	if err != errDummy {
		t.Error(err)
	}
}

func TestEventualSuccess(t *testing.T) {
	ctx := context.Background()
	g := gosyphus.New(1*time.Millisecond, 30*time.Millisecond)
	n := 0
	err := g.Do(ctx, func() error {
		if n < 1 {
			n++
			return errDummy
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}

func TestEventualFailure(t *testing.T) {
	ctx := context.Background()
	g := gosyphus.New(1*time.Millisecond, 30*time.Millisecond)
	n := 0
	err := g.Dos(ctx, func() error {
		if n < 1 {
			n++
			return errDummy
		}
		return errUnrecoverable
	}, func(err error) bool {
		return err != errUnrecoverable
	})
	if err != errUnrecoverable {
		t.Error(err)
	}
}

func TestBadConfig(t *testing.T) {
	ctx := context.Background()
	g := gosyphus.New(0, -1)
	n := 0
	err := g.Do(ctx, func() error {
		if n < 1 {
			n++
			return errDummy
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
