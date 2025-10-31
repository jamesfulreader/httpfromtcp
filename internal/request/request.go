package request

import (
	"bufio"
	"errors"
	"io"
	"strings"
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
	bufReader := bufio.NewReader(reader)
	reqLineStr, err := bufReader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	parts := strings.Fields(reqLineStr)
	if len(parts) < 3 {
		return nil, errors.New("error parsing req")
	}

	method := parts[0]
	reqTarget := parts[1]
	httpVer := parts[2]

	if httpVer != "HTTP/1.1" {
		return nil, errors.New("reqlineStr error")
	}

	version := strings.TrimPrefix(httpVer, "HTTP/")

	return &Request{
		RequestLine: RequestLine{
			HttpVersion:   version,
			RequestTarget: reqTarget,
			Method:        strings.ToUpper(method),
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
