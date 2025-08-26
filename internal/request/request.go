package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers

	state requestState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type requestState int

const (
	requestStateInitialised requestState = iota
	requestStateParsingHeaders
	requestStateDone
)

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state: requestStateInitialised,
	}
	for req.state != requestStateDone {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}
		numBytesRead, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state != requestStateDone {
					return nil, fmt.Errorf("incomplete request")
				}
				break
			}
			return nil, err
		}
		readToIndex += numBytesRead

		numBytesParsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[numBytesParsed:])
		readToIndex -= numBytesParsed
	}

	return req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.state {
	case requestStateInitialised:
		requestLine, n, err := parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return 0, nil
		}
		r.RequestLine = *requestLine
		r.state = requestStateParsingHeaders
		return n, nil
	case requestStateParsingHeaders:
		requestHeader := parseRequestHeaders(data)
	case requestStateDone:
		return 0, fmt.Errorf("error: trying to read data in a done state")
	default:
		return 0, fmt.Errorf("unkown state")
	}
}

func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, []byte(crlf))
	if idx == -1 {
		return nil, 0, nil
	}
	parts := bytes.Split(b, []byte(crlf))
	requestLine, err := constructRequestLine(parts)
	if err != nil {
		return nil, 0, err
	}
	return requestLine, idx + 2, nil
}

func parseRequestHeaders(b []byte) (*headers.Headers, int, error) {

}

func constructRequestLine(parts [][]byte) (*RequestLine, error) {
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
