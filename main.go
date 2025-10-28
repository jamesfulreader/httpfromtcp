package main

import (
	"fmt"
	"io"
	"log"
	"os"
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

	file, err := os.Open("./message.txt")
	if err != nil {
		log.Fatal(err)
	}

	lines := getLinesChannel(file)
	for line := range lines {
		fmt.Printf("read: %s\n", line)
	}
}
