# ya9p
[![Go Reference](https://pkg.go.dev/badge/github.com/rmatsuoka/ya9p.svg)](https://pkg.go.dev/github.com/rmatsuoka/ya9p)

**This package is experimental.**

Package ya9p provides 9P server implementations.
This package provides only the minimum functionality required to serve 9P.
In addition, it can serve filesystems defined in fs.FS.

# example
Let's serve the local file system with 9P.
``` Go
package main

import (
	"os"
	"net"
	"log"

	"github.com/rmatsuoka/ya9p"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}
		go ya9p.ServeFS(conn, os.DirFS("/"))
	}
}
```
