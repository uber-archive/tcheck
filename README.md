# tcheck

A simple TChannel health checker written in Go.

`tcheck` supports the following flags:

* `--peer` a singular host:port to health check
* `--serviceName` the target's service name

Examples:

```
tcheck --peer 127.0.0.1:4532 --serviceName keyvalue
```

## Tests

Run tests using `go test`.

## License

MIT.
