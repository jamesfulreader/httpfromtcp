package main

import (
	"io"
	"log"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go func(c net.Conn) {
			defer c.Close()
			_, err := io.Copy(os.Stdout, c)
			if err != nil {
				log.Printf("error copying connection to stdout: %v", err)
			}
		}(conn)
	}
}
