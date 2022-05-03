package main

import (
	"log"
	"net"
	"os"

	"github.com/rmatsuoka/ya9p"
)

func main() {
	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	var listener net.Listener
	listener, err = net.Listen("tcp", "0.0.0.0:5640")
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("connected from %v", conn.RemoteAddr())
		go ya9p.Serve(conn, ya9p.FS(os.DirFS(homedir)))
	}
}
