package request

import (
	"errors"
	"fmt"
	"io"
	"strconv"
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

// parseRequestBlock looks for the end of headers ("\r\n\r\n") and
// returns the request-line parts and the total bytes that should be
// consumed (headers + body if Content-Length present). If more data
// is needed it returns bytesConsumed==0.
func parseRequestBlock(data []byte) ([]string, int, error) {
	s := string(data)

	headersEnd := strings.Index(s, "\r\n\r\n")
	if headersEnd == -1 {
		return nil, 0, nil // need more data
	}

	// headers block includes the trailing CRLFCRLF
	headersBlock := s[:headersEnd+4]
	lines := strings.Split(headersBlock, "\r\n")
	if len(lines) == 0 {
		return nil, 0, errors.New("empty request")
	}

	parts := strings.Split(lines[0], " ")
	if len(parts) != 3 {
		return nil, 0, errors.New("invalid request line format")
	}

	// look for Content-Length header (case-insensitive)
	contentLength := 0
	for _, hdr := range lines[1:] {
		if hdr == "" {
			continue
		}
		colonIdx := strings.Index(hdr, ":")
		if colonIdx == -1 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(hdr[:colonIdx]))
		val := strings.TrimSpace(hdr[colonIdx+1:])
		if name == "content-length" {
			n, err := strconv.Atoi(val)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid content-length: %v", err)
			}
			contentLength = n
		}
	}

	totalNeeded := headersEnd + 4 + contentLength
	if len(data) < totalNeeded {
		return nil, 0, nil // need more data
	}

	return parts, totalNeeded, nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.state == stateInitialized {
		parts, bytesConsumed, err := parseRequestBlock(data)
		if err != nil {
			return 0, err
		}
		if bytesConsumed == 0 {
			return 0, nil // more data needed but not an error
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
