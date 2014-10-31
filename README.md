## streamon

Exec command output stream consumer parser written in Go.

### Activating debug messages

streamon package debug logging output can be enabled through the use of `-ldflags` when building or running your program.

Example usage:

    go run -ldflags '-X streamon.debugEnabled true' *.go

