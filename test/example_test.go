package test

import (
	"fmt"
	"time"

	"github.com/lufia/mock"
)

func Example() {
	now := time.Date(2024, time.April, 1, 10, 12, 50, 0, time.UTC)
	mock.Set(time.Now, func() time.Time {
		return now
	})
	fmt.Println(time.Now().Format(time.RFC3339))
	// Output: 2024-04-01T10:12:50Z
}
