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

	requestMethod := string(requestLineParts[0])
	var methodRe = regexp.MustCompile(`^[A-Z]+$`)
	if !methodRe.MatchString(requestMethod) || !isValidMethod(requestMethod) {
		return RequestLine{}, fmt.Errorf("request method is not valid: %s", requestMethod)
	}

	httpVersion := string(bytes.Split(requestLineParts[2], []byte("/"))[1])
	if httpVersion != "1.1" {
		return RequestLine{}, fmt.Errorf("unsupported http version used, this service only supports http 1.1")
	}

	requestTarget := string(requestLineParts[1])

	requestLine := RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: requestTarget,
		Method:        requestMethod,
	}
	return requestLine, nil
}

func isValidMethod(requestMethod string) bool {
	var allowedMethods = map[string]struct{}{
		"GET":     {},
		"HEAD":    {},
		"POST":    {},
		"PUT":     {},
		"DELETE":  {},
		"CONNECT": {},
		"OPTIONS": {},
		"TRACE":   {},
		"PATCH":   {},
	}
	_, ok := allowedMethods[requestMethod]
	return ok
}
