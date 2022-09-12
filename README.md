# no-load

A proxy to minimize real requests performed. Every request is identified by request method, path, headers and body.

Usage:

```shell
no-load <target-url> <data-dir> [--ignore-header="Authorization"] [--ignore-header="Example2"] [...]
```

Ignored headers are not respected when hashing the request.
