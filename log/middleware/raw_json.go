package middleware

import (
	"encoding"
	"encoding/json"
	"strings"
)

var emptyString = []byte(`""`)

// rawData represent a raw json string.
type rawData []byte

var (
	_ json.Marshaler         = rawData(nil)
	_ encoding.TextMarshaler = rawData(nil)
)

// MarshalJSON implements the interface json.Marshaler, which will returns
// the original string without the leading and trailing whitespace as []byte
// if not empty, else []byte(`""`) instead.
func (s rawData) MarshalJSON() ([]byte, error) {
	if js := strings.TrimSpace(string(s)); len(js) > 0 {
		return []byte(js), nil
	}
	return emptyString, nil
}

func (s rawData) MarshalText() ([]byte, error) {
	js, err := s.compact()
	return []byte(js), err
}

func (s rawData) compact() (js string, err error) {
	if js = strings.TrimSpace(string(s)); len(js) > 0 {
		buf := getbuffer()
		if err = json.Compact(buf, []byte(js)); err == nil {
			js = buf.String()
		}
		putbuffer(buf)
	}
	return
}
