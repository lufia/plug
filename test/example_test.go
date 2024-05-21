package test

import (
	"fmt"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/lufia/mock"
)

func Example_timeNow() {
	scope := mock.CurrentScope()
	defer scope.Delete()

	now := time.Date(2024, time.April, 1, 10, 12, 50, 0, time.UTC)
	mock.Set(time.Now, func() time.Time {
		return now
	})
	fmt.Println(time.Now().Format(time.RFC3339))
	// Output: 2024-04-01T10:12:50Z
}

func Example_osUserCurrent() {
	scope := mock.CurrentScope()
	defer scope.Delete()

	mock.Set(user.Current, func() (*user.User, error) {
		return &user.User{Uid: "100", Username: "mock"}, nil
	})
	u, _ := user.Current()
	fmt.Println(u.Username, u.Uid)
	// Output: mock 100
}

func Example_netHttpClientDo() {
	scope := mock.CurrentScope()
	defer scope.Delete()

	mock.Set((*http.Client)(nil).Do, func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	})
	resp, _ := http.Get("https://example.com")
	fmt.Println(resp.StatusCode)
	// Output: 200
}

func Example_osGetpid() {
	scope := mock.CurrentScope()
	defer scope.Delete()

	mock.Set(os.Getpid, func() int {
		return 1
	})
	fmt.Println(os.Getpid())
	// Output: 1
}
