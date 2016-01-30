# tcheck

Simple go TChannel / Hyperbahn health checker.

Takes these flags:

* `--hostsFile` specify Hyperbahn hosts file; default is /etc/uber/hyperbahn/hosts.json
* `--peer` a singular host:port to hit; overrides `--hostsFile`
* `--serviceName` specify service name; default hyperbahn

Examples:

```
tcheck --peer 127.0.0.1:21300 --serviceName populous
tcheck --serviceName keyvalue
```

## Tests

Tests use `exec` so you'll have to `go build` first, then `go test`.

## License

MIT.
