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
