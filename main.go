package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	fmt.Println("starting to read the file")

	content, err := os.Open("./message.txt")
	if err != nil {
		log.Fatal(err)
	}
	myByte := make([]byte, 8)
	for {
		conentByte, err := content.Read(myByte)
		if err != nil {
			if err == io.EOF {
				fmt.Println("End of file reached")
				break
			}
			break
		}
		fmt.Printf("read: %s\n", string(myByte[:conentByte]))
	}
}
