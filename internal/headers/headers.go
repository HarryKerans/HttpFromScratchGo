package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

// Set sets the header key to the given value.
func (h Headers) Set(key, value string) {
	h[key] = value
}

const crlf = "\r\n"

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	} else if idx == 0 {
		return 2, true, nil
	}

	headerFieldName, headerFieldValue, err := retrieveHeaderParts(data, idx, byte(':'))
	if err != nil {
		return 0, false, err
	}

	h.Set(headerFieldName, headerFieldValue)
	return idx + 2, false, nil
}

func retrieveHeaderParts(data []byte, idx int, char byte) (headerFieldName, headerFieldValue string, err error) {
	parts := bytes.SplitN(data[:idx], []byte(":"), 2)

	method := parts[0]
	for _, c := range method {
		if (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') && (c < 0 || c > 9) && (strings.IndexByte("!#$%&'*+-.^_`|~", c) == -1) && (c != ' ') {
			return "", "", fmt.Errorf("invalid characters in header field-name")
		}
	}

	headerFieldName = strings.ToLower(string(parts[0]))
	headerFieldValue = string(parts[1])

	if headerFieldName[len(headerFieldName)-1] == ' ' {
		return "", "", fmt.Errorf("incorrect header name format, trailing whitespace before the \":\" seperator")
	}

	headerFieldName = strings.Trim(headerFieldName, " ")
	headerFieldValue = strings.Trim(headerFieldValue, " ")

	return headerFieldName, headerFieldValue, nil
}
