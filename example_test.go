package plug_test

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"os/user"
	"time"

	"github.com/lufia/plug"
)

func Example_timeNow() {
	scope := plug.CurrentScope()
	defer scope.Delete()

	now := time.Date(2024, time.April, 1, 10, 12, 50, 0, time.UTC)
	key := plug.Func("time.Now", time.Now)
	plug.Set(key, func() time.Time {
		return now
	})
	fmt.Println(time.Now().Format(time.RFC3339))
	// Output: 2024-04-01T10:12:50Z
}

func Example_osUserCurrent() {
	scope := plug.CurrentScope()
	defer scope.Delete()

	key := plug.Func("os/user.Current", user.Current)
	plug.Set(key, func() (*user.User, error) {
		return &user.User{Uid: "100", Username: "user"}, nil
	})
	u, _ := user.Current()
	fmt.Println(u.Username, u.Uid)
	// Output: user 100
}

func Example_netHttpClientDo() {
	scope := plug.CurrentScope()
	defer scope.Delete()

	key := plug.Func("net/http.Client.Do", (*http.Client)(nil).Do)
	plug.Set(key, func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200}, nil
	})
	resp, _ := http.Get("https://example.com")
	fmt.Println(resp.StatusCode)
	// Output: 200
}

func Example_osGetpid() {
	scope := plug.CurrentScope()
	defer scope.Delete()

	key := plug.Func("os.Getpid", os.Getpid)
	plug.Set(key, func() int {
		return 1
	})
	fmt.Println(os.Getpid())
	// Output: 1
}

func Example_mathRandV2N() {
	scope := plug.CurrentScope()
	defer scope.Delete()

	key := plug.Func("math/rand/v2.N", rand.N[int])
	plug.Set(key, func(n int) int {
		return 3
	})
	fmt.Println(rand.N[int](10))
	// Output: 3
}
