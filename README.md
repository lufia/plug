# mock

## Usage

```sh
go test -overlay <(go run github.com/lufia/mock/cmd/mock@latest)
```

## Limitation

* cyclic import
* `go:linkname`
