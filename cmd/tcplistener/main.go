package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		defer f.Close()

		buf := make([]byte, 8)
		currentLine := ""

		for {
			n, err := f.Read(buf)
			if n > 0 {
				seg := string(buf[:n])
				parts := strings.Split(seg, "\n")

				for i := 0; i < len(parts)-1; i++ {
					line := currentLine + parts[i]
					ch <- line
					currentLine = ""
				}

				currentLine += parts[len(parts)-1]
			}

			if err != nil {
				if err == io.EOF {
					if len(currentLine) > 0 {
						ch <- currentLine
					}
					return
				}
				log.Fatal(err)
			}
		}
	}()

	return ch
}

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

		fmt.Println("tcp connection on PORT 42069 accepted")

		go func(c net.Conn) {
			lines := getLinesChannel(c)
			for line := range lines {
				fmt.Println(line)
			}
			c.Close()
		}(conn)
		fmt.Println("tcp connection closed")
	}

}
