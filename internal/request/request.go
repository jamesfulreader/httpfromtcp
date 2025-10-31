package request

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	request, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	parts, err := parseRequestLine(request)
	if err != nil {
		return nil, err
	}

	method := parts[0]
	requestTarget := parts[1]
	httpVersion := parts[2]

	for _, char := range method {
		if !unicode.IsUpper(char) || !unicode.IsLetter(char) {
			return nil, errors.New("invalid method")
		}
	}

	if httpVersion != "HTTP/1.1" {
		return nil, errors.New("invalid HTTP version")
	}

	version := strings.TrimPrefix(httpVersion, "HTTP/")

	return &Request{
		RequestLine: RequestLine{
			HttpVersion:   version,
			RequestTarget: requestTarget,
			Method:        method,
		},
	}, nil
}

func parseRequestLine(request []byte) ([]string, error) {
	s := string(request)
	// get the first line (request-line)
	lines := strings.SplitN(s, "\r\n", 2)
	first := lines[0]
	parts := strings.Fields(first)
	if len(parts) >= 3 {
		// Request-Line = Method SP Request-Target SP HTTP-Version
		return parts, nil
	}
	return parts, errors.New("error parsing request")
}
