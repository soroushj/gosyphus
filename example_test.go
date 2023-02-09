package gosyphus_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/soroushj/gosyphus"
)

func Example() {
	n := 0
	printTheAnswer := func() error {
		fmt.Println("calculating the answer")
		if n < 2 {
			n++
			return errors.New("calculation error")
		}
		fmt.Println("the answer is 42")
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	err := gosyphus.Do(ctx, printTheAnswer)
	if err != nil {
		fmt.Println("failure:", err)
		return
	}
	fmt.Println("success")
	// Output: calculating the answer
	// calculating the answer
	// calculating the answer
	// the answer is 42
	// success
}

func Example_shouldRetry() {
	impossible := errors.New("impossible to calculate the answer")
	n := 0
	printTheAnswer := func() error {
		fmt.Println("calculating the answer")
		if n < 2 {
			n++
			return errors.New("calculation error")
		}
		return impossible
	}
	shouldRetry := func(err error) bool {
		return err != impossible
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()
	err := gosyphus.Dos(ctx, printTheAnswer, shouldRetry)
	if err != nil {
		fmt.Println("failure:", err)
		return
	}
	fmt.Println("success")
	// Output: calculating the answer
	// calculating the answer
	// calculating the answer
	// failure: impossible to calculate the answer
}

func Example_timeout() {
	printTheAnswer := func() error {
		return errors.New("calculation error")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	g := gosyphus.New(100*time.Millisecond, 500*time.Millisecond)
	err := g.Do(ctx, printTheAnswer)
	if err != nil {
		fmt.Println("failure:", err)
		return
	}
	fmt.Println("success")
	// Output: failure: context deadline exceeded
}
