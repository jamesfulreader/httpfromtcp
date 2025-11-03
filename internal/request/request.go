package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
	state       int
}

const (
	stateInitialized = iota
	stateDone
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := &Request{
		state: stateInitialized,
	}
	buffer := make([]byte, 8)
	bytesReadFromReader := 0

	for {
		n, err := reader.Read(buffer[bytesReadFromReader:])
		if err != nil {
			if err == io.EOF {
				if request.state != stateDone {
					return nil, errors.New("incomplete request")
				}
				break
			}
			return nil, fmt.Errorf("error reading: %v", err)
		}

		bytesReadFromReader += n

		bytesConsumed, err := request.parse(buffer[0:bytesReadFromReader])
		if err != nil {
			return nil, fmt.Errorf("error with parse: %v", err)
		}

		remainingBytes := bytesReadFromReader - bytesConsumed
		if remainingBytes > 0 {
			copy(buffer[0:], buffer[bytesConsumed:bytesReadFromReader])
		}
		bytesReadFromReader = remainingBytes // after shift set bytesReadFromReader to remainingBytes

		if request.state == stateDone {
			break
		}

		if bytesReadFromReader >= len(buffer) {
			doubleBuffer := make([]byte, len(buffer)*2)
			copy(doubleBuffer[0:], buffer[0:])
			buffer = doubleBuffer
		}
	}
	return request, nil
}

func parseRequestLine(request []byte) ([]string, int) {
	s := string(request)

	indexRN := strings.Index(s, "\r\n")

	if !strings.Contains(s, "\r\n") {
		return nil, 0
	} else {
		bytesConsumed := indexRN + len("\r\n")
		requestLine := strings.Split(s, "\r\n")
		parts := strings.Split(requestLine[0], " ")

		if len(parts) != 3 {
			return nil, 0
		}
		return parts, bytesConsumed
	}
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == stateInitialized {
		parts, bytesConsumed := parseRequestLine(data)
		if bytesConsumed == 0 {
			return 0, nil // more data needed but not an error
		}

		if len(parts) != 3 {
			return bytesConsumed, errors.New("invalid request line format")
		}

		method := parts[0]
		requestTarget := parts[1]
		httpVersion := parts[2]

		if method == "" {
			return bytesConsumed, errors.New("invalid method: method cannot be empty")
		}

		for _, char := range method {
			if !unicode.IsUpper(char) || !unicode.IsLetter(char) {
				return bytesConsumed, errors.New("invalid method")
			}
		}

		if httpVersion != "HTTP/1.1" {
			return bytesConsumed, errors.New("invalid HTTP version")
		}

		r.RequestLine = RequestLine{
			Method:        method,
			RequestTarget: requestTarget,
			HttpVersion:   strings.TrimPrefix(httpVersion, "HTTP/"), // store just 1.1
		}

		r.state = stateDone
		return bytesConsumed, nil
	}
	return 0, nil
}
