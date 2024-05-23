# plug

Package plug replaces functions, defined in the other packages, on testing.

[![GoDev][godev-image]][godev-url]

## Usage

See [examples](https://github.com/lufia/plug/blob/main/test/example_test.go).

```sh
go test -overlay <(go run github.com/lufia/plug/cmd/plug@latest)
```

Then add below to **.gitignore**

```txt
plug
```

## Limitations

* cyclic import: runtime, reflect, etc.
* **go:linkname** functions: *time.Sleep*, etc.

[godev-image]: https://pkg.go.dev/badge/github.com/lufia/plug
[godev-url]: https://pkg.go.dev/github.com/lufia/plug
