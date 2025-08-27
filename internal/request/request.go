package request

import (
	"bytes"
	"errors"
	"fmt"
	"httpfromtcp/internal/headers"
	"io"
	"regexp"
	"strconv"
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       requestState
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
	requestStateParsingBody
	requestStateDone
)

const crlf = "\r\n"
const bufferSize = 8

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0
	req := &Request{
		state:   requestStateInitialised,
		Headers: headers.NewHeaders(),
		Body:    make([]byte, 0),
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
					return nil, fmt.Errorf("incomplete request, in state: %d, read n bytes on EOF: %d", req.state, numBytesRead)
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

	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}
	return totalBytesParsed, nil
}

func (r *Request) parseSingle(b []byte) (int, error) {
	switch r.state {
	case requestStateInitialised:
		requestLine, n, err := parseRequestLine(b)
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
		n, done, err := r.Headers.Parse(b)
		if err != nil {
			return 0, err
		}
		if done {
			r.state = requestStateParsingBody
		}
		return n, nil
	case requestStateParsingBody:
		contentLengthStr, ok := r.Headers.Get("content-length")
		if !ok {
			r.state = requestStateDone
			return len(b), nil
		}
		contentLen, err := strconv.Atoi(contentLengthStr)
		if err != nil {
			return 0, fmt.Errorf("Malformed content-length: %s", err)
		}
		r.Body = append(r.Body, b...)
		if len(r.Body) > contentLen {
			return 0, fmt.Errorf("body length exceeds stated content-length")
		}
		if len(r.Body) == contentLen {
			r.state = requestStateDone
		}
		return len(b), nil
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
