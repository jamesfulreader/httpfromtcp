package main

import (
	"fmt"
	"log"
	"net"

	"github.com/jamesfulreader/httpfromtcp/internal/request"
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
			// _, err := io.Copy(os.Stdout, c)
			// if err != nil {
			// 	log.Printf("error copying connection to stdout: %v", err)
			// }
			req, err := request.RequestFromReader(c)
			if err != nil {
				log.Printf("error parsing request: %v", err)
			}
			fmt.Printf(
				"Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
				req.RequestLine.Method,
				req.RequestLine.RequestTarget,
				req.RequestLine.HttpVersion)
		}(conn)
	}
}
