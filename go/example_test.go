package leaksnow_test

import (
	"context"
	"fmt"
	"time"

	leaksnow "github.com/ItsMariusBC/leaksnow-wrapper/go"
)

func ExampleClient_Search() {
	c := leaksnow.NewClient("ms_xxx",
		leaksnow.WithTimeout(15*time.Second),
		leaksnow.WithRetry(leaksnow.RetryConfig{
			MaxRetries: 3,
			BaseDelay:  500 * time.Millisecond,
			MaxDelay:   8 * time.Second,
			RetryOn:    []int{429, 500, 502, 503, 504},
		}),
	)
	_ = c
	fmt.Println("client ready")
	// Output: client ready
}

func ExampleClient_Search_call() {
	// Demonstrates the call shape; not executed against a live server here.
	run := func(c *leaksnow.Client) error {
		ctx := context.Background()
		if _, err := c.Search(ctx, leaksnow.SearchRequest{
			Query:    "host:example.com",
			Scope:    leaksnow.ScopeLeak,
			Severity: leaksnow.SeverityAll,
		}); err != nil {
			return err
		}
		return nil
	}
	_ = run
	fmt.Println("ok")
	// Output: ok
}
