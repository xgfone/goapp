package middleware

import (
	"encoding"
	"encoding/json"
	"strings"
)

var emptyString = []byte(`""`)

// rawString represent a raw json string.
type rawString string

var (
	_ json.Marshaler         = rawString("")
	_ encoding.TextMarshaler = rawString("")
)

// MarshalJSON implements the interface json.Marshaler, which will returns
// the original string without the leading and trailing whitespace as []byte
// if not empty, else []byte(`""`) instead.
func (s rawString) MarshalJSON() ([]byte, error) {
	if js := strings.TrimSpace(string(s)); len(js) > 0 {
		return []byte(js), nil
	}
	return emptyString, nil
}

func (s rawString) MarshalText() ([]byte, error) {
	js, err := s.compact()
	return []byte(js), err
}

func (s rawString) compact() (js string, err error) {
	if js = strings.TrimSpace(string(s)); len(js) > 0 {
		buf := getbuffer()
		if err = json.Compact(buf, []byte(js)); err == nil {
			js = buf.String()
		}
		putbuffer(buf)
	}
	return
}
