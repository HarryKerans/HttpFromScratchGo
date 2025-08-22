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

const crlf = "\r\n"

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
		RequestLine: *requestLine,
	}
	return request, nil
}

func parseRequestLine(b []byte) (*RequestLine, error) {
	idx := bytes.Index(b, []byte(crlf))
	if idx == -1 {
		return nil, fmt.Errorf("could not find CRLF in request-line")
	}
	parts := bytes.Split(b, []byte(crlf))

	requestLineParts := bytes.Split(parts[0], []byte(" "))

	if len(requestLineParts) != 3 {
		return nil, fmt.Errorf("malformed request line: %s", string(parts[0]))
	}

	requestMethod := string(requestLineParts[0])
	var methodRe = regexp.MustCompile(`^[A-Z]+$`)
	if !methodRe.MatchString(requestMethod) || !isValidMethod(requestMethod) {
		return nil, fmt.Errorf("request method is not valid: %s", requestMethod)
	}
	httpVersionAndProtocol := bytes.Split(requestLineParts[2], []byte("/"))
	httpProtocol := string(httpVersionAndProtocol[0])
	httpVersion := string(httpVersionAndProtocol[1])
	if httpProtocol != "HTTP" {
		return nil, fmt.Errorf("unsupported protocol, this service is designed for HTTP")
	}
	if httpVersion != "1.1" {
		return nil, fmt.Errorf("unsupported http version used, this service only supports http 1.1")
	}

	requestTarget := string(requestLineParts[1])

	requestLine := &RequestLine{
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
