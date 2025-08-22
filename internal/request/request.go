package request

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
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
	b, err := io.ReadAll(reader)
	if err != nil {
		return &Request{}, err
	}
	requestLine, err := parseRequestLine(b)
	if err != nil {
		return &Request{}, err
	}
	request := &Request{
		RequestLine: requestLine,
	}
	return request, nil

}

func parseRequestLine(b []byte) (RequestLine, error) {
	parts := bytes.Split(b, []byte("\r\n"))

	requestLineParts := bytes.Split(parts[0], []byte(" "))

	if len(requestLineParts) != 3 {
		return RequestLine{}, fmt.Errorf("malformed request line: %s", string(parts[0]))
	}

	var methodRe = regexp.MustCompile(`^[A-Z]+$`)
	if !methodRe.MatchString(string(requestLineParts[0])) {
		return RequestLine{}, fmt.Errorf("request method is not valid: %s", string(requestLineParts[0]))
	}
	httpVersion := string(bytes.Split(requestLineParts[2], []byte("/"))[1])
	if httpVersion != "1.1" {
		return RequestLine{}, fmt.Errorf("unsupported http version used, this service only supports http 1.1")
	}

	requestLine := RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: string(requestLineParts[1]),
		Method:        string(requestLineParts[0]),
	}
	return requestLine, nil
}
