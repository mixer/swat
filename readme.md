# swat [![GoDoc](https://godoc.org/github.com/WatchBeam/swat?status.svg)](https://godoc.org/github.com/WatchBeam/swat) [![Build Status](https://travis-ci.org/WatchBeam/swat.svg)](https://travis-ci.org/WatchBeam/swat)

A general-purpose tool for debugging and analyzing programs, in development and production. Example:

```go
package main

import (
    "os/syscall"
    "os"
    "time"
    "github.com/WatchBeam/swat"
)

func main() {
    s := swat.Start(
        swat.DumpGoroutine().
            OnSignal(syscall.SIGUSR1).
            ToWriter(os.Stdout),
        swat.DumpHeap().
            After(5*time.Minute).
            Every(time.Second).
            For(60*time.Second).
            ToFile("record.csv"),
    )
    defer s.End()

    // do the necessary!
}

```
