// Copyright 2020~2022 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package validate supplies a struct validator to validate whether the field
// value of the struct is valid or not.
package validate

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/xgfone/cast"
	"github.com/xgfone/go-apiserver/http/binder"
	"github.com/xgfone/go-apiserver/http/reqresp"
)

// Validate is the default global validator.
var Validate = validator.New()

// ValidateErrorFormatters is the set of formatters to format validator.FieldError.
var ValidateErrorFormatters = make(map[string]func(validator.FieldError) error)

// RegisterValidateErrorFormatter registers the field error formatter
// with the name, which will override it if the name has been registered.
//
// If name is empty, it will panic. If f is nil, it will delete it.
func RegisterValidateErrorFormatter(name string, f func(validator.FieldError) error) {
	if name == "" {
		panic("name is empty")
	}

	if f == nil {
		delete(ValidateErrorFormatters, name)
	} else {
		ValidateErrorFormatters[name] = f
	}
}

// SetValidateTagName registers a tag name function with the given tags into v.
func SetValidateTagName(validator *validator.Validate, tags ...string) {
	if len(tags) == 0 {
		return
	}

	validator.RegisterTagNameFunc(func(field reflect.StructField) string {
		for _, tag := range tags {
			if v := strings.TrimSpace(field.Tag.Get(tag)); v != "" {
				return v
			}
		}
		return field.Name
	})
}

// StructValidator returns a new struct validator function, which is used to
// set the validator of the github.com/xgfone/ship#Ship.Validator.
//
// If validate is nil, it is the global Validate by default.
func StructValidator(validate *validator.Validate) func(interface{}) error {
	if validate == nil {
		validate = Validate
	}

	return func(v interface{}) (err error) {
		if err = validate.Struct(v); err == nil {
			return
		} else if errs, ok := err.(validator.ValidationErrors); ok {
			es := make([]string, len(errs))
			for i, e := range errs {
				if f, ok := ValidateErrorFormatters[e.Tag()]; ok && f != nil {
					es[i] = f(e).Error()
				} else {
					es[i] = e.(error).Error()
				}
			}
			return errors.New(strings.Join(es, ", "))
		}

		return
	}
}

func init() {
	reqresp.DefaultBinder.(*binder.DefaultValidateBinder).Validate = StructValidator(nil)

	SetValidateTagName(Validate, "json", "query")
	Validate.RegisterValidation("addr", func(fl validator.FieldLevel) bool {
		host, port, err := net.SplitHostPort(fl.Field().String())
		return host != "" && port != "" && err == nil
	})
	Validate.RegisterValidation("zero", func(fl validator.FieldLevel) bool {
		return cast.IsZero(fl.Field().Interface())
	})
	Validate.RegisterValidation("notzero", func(fl validator.FieldLevel) bool {
		return !cast.IsZero(fl.Field().Interface())
	})

	// Zero
	RegisterValidateErrorFormatter("zero", func(fe validator.FieldError) error {
		return fmt.Errorf("'%s' is not zero", fe.Field())
	})

	// NotZero
	RegisterValidateErrorFormatter("notzero", func(fe validator.FieldError) error {
		return fmt.Errorf("'%s' is zero", fe.Field())
	})

	// Addr
	RegisterValidateErrorFormatter("addr", func(fe validator.FieldError) error {
		return fmt.Errorf("invalid addr value '%v' of '%s'", fe.Value(), fe.Field())
	})

	// One Of
	RegisterValidateErrorFormatter("oneof", func(fe validator.FieldError) error {
		return fmt.Errorf("invalid value '%v' of '%s' is not in [%s]",
			fe.Value(), fe.Field(), fe.Param())
	})

	// Required
	RegisterValidateErrorFormatter("required", func(fe validator.FieldError) error {
		return fmt.Errorf("missing '%s'", fe.Field())
	})

	// Min
	RegisterValidateErrorFormatter("min", func(fe validator.FieldError) error {
		switch reflect.TypeOf(fe.Value()).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return fmt.Errorf("the length of '%s' is less than %s", fe.Field(), fe.Param())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Errorf("invalid value '%v' of '%s': less than %s",
				fe.Value(), fe.Field(), fe.Param())
		default:
			return fe.(error)
		}
	})

	// Max
	RegisterValidateErrorFormatter("max", func(fe validator.FieldError) error {
		switch reflect.TypeOf(fe.Value()).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return fmt.Errorf("the length of '%s' is greater than %s", fe.Field(), fe.Param())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Errorf("invalid value '%v' of '%s': greater than %s",
				fe.Value(), fe.Field(), fe.Param())
		default:
			return fe.(error)
		}
	})

	// Greater Than
	RegisterValidateErrorFormatter("gt", func(fe validator.FieldError) error {
		switch reflect.TypeOf(fe.Value()).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return fmt.Errorf("the length of '%s' is not greater than %s", fe.Field(), fe.Param())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Errorf("invalid value '%v' of '%s': not greater than %s",
				fe.Value(), fe.Field(), fe.Param())
		default:
			return fe.(error)
		}
	})

	// Greater Than or Equal
	RegisterValidateErrorFormatter("gte", func(fe validator.FieldError) error {
		switch reflect.TypeOf(fe.Value()).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return fmt.Errorf("the length of '%s' is not greater than or equal to %s",
				fe.Field(), fe.Param())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Errorf("invalid value '%v' of '%s': not greater than or equal to %s",
				fe.Value(), fe.Field(), fe.Param())
		default:
			return fe.(error)
		}
	})

	// Less Than
	RegisterValidateErrorFormatter("lt", func(fe validator.FieldError) error {
		switch reflect.TypeOf(fe.Value()).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return fmt.Errorf("the length of '%s' is not less than %s", fe.Field(), fe.Param())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Errorf("invalid value '%v' of '%s': not less than %s",
				fe.Value(), fe.Field(), fe.Param())
		default:
			return fe.(error)
		}
	})

	// Less Than or Equal
	RegisterValidateErrorFormatter("lte", func(fe validator.FieldError) error {
		switch reflect.TypeOf(fe.Value()).Kind() {
		case reflect.Slice, reflect.Array, reflect.Map, reflect.String:
			return fmt.Errorf("the length of '%s' is not less than or equal to %s",
				fe.Field(), fe.Param())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64:
			return fmt.Errorf("invalid value '%v' of '%s': not less than or equal to %s",
				fe.Value(), fe.Field(), fe.Param())
		default:
			return fe.(error)
		}
	})

	// Alpha
	RegisterValidateErrorFormatter("alpha", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not aplha", fe.Field())
	})

	// Alphanumeric
	RegisterValidateErrorFormatter("alphanumeric", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not aplhanumber", fe.Field())
	})

	// Alpha Unicode
	RegisterValidateErrorFormatter("alphaunicode", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not aplha unicode", fe.Field())
	})

	// Alphanumeric Unicode
	RegisterValidateErrorFormatter("alphanumunicode", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not aplhanumeric unicode", fe.Field())
	})

	// Number
	RegisterValidateErrorFormatter("number", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not number", fe.Field())
	})

	// Numeric
	RegisterValidateErrorFormatter("numeric", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not numeric", fe.Field())
	})

	// Hexadecimal String
	RegisterValidateErrorFormatter("hexadecimal", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not hexadecimal", fe.Field())
	})

	// Lowercase String
	RegisterValidateErrorFormatter("lowercase", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not lowercase", fe.Field())
	})

	// Uppercase String
	RegisterValidateErrorFormatter("uppercase", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not uppercase", fe.Field())
	})

	// E-mail String
	RegisterValidateErrorFormatter("email", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not email", fe.Field())
	})

	// JSON String
	RegisterValidateErrorFormatter("json", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not json", fe.Field())
	})

	// File path
	RegisterValidateErrorFormatter("file", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not file path", fe.Field())
	})

	// URL String
	RegisterValidateErrorFormatter("url", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not url", fe.Field())
	})

	// URI String
	RegisterValidateErrorFormatter("uri", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not uri", fe.Field())
	})

	// Base64 String
	RegisterValidateErrorFormatter("base64", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not base64", fe.Field())
	})

	// Base64URL String
	RegisterValidateErrorFormatter("base64url", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not base64 url", fe.Field())
	})

	// Contains
	RegisterValidateErrorFormatter("contains", func(fe validator.FieldError) error {
		return fmt.Errorf("the value '%v' of '%s' does not contain '%s'",
			fe.Value(), fe.Field(), fe.Param())
	})

	// Universally Unique Identifier UUID
	RegisterValidateErrorFormatter("uuid", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not uuid", fe.Field())
	})

	// Universally Unique Identifier UUID
	RegisterValidateErrorFormatter("uuid4", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not uuid4", fe.Field())
	})

	// ASCII
	RegisterValidateErrorFormatter("ascii", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not ascii", fe.Field())
	})

	// Printable ASCII
	RegisterValidateErrorFormatter("printascii", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not printable ascii", fe.Field())
	})

	// Data URL
	RegisterValidateErrorFormatter("datauri", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not datauri", fe.Field())
	})

	// Internet Protocol Address IP
	RegisterValidateErrorFormatter("ip", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not ip", fe.Field())
	})

	// Internet Protocol Address IPv4
	RegisterValidateErrorFormatter("ipv4", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not ipv4", fe.Field())
	})

	// Internet Protocol Address IPv6
	RegisterValidateErrorFormatter("ipv6", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not ipv6", fe.Field())
	})

	// Classless Inter-Domain Routing CIDR
	RegisterValidateErrorFormatter("cidr", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not CIDR", fe.Field())
	})

	// Classless Inter-Domain Routing CIDRv4
	RegisterValidateErrorFormatter("cidrv4", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not CIDRv4", fe.Field())
	})

	// Classless Inter-Domain Routing CIDRv6
	RegisterValidateErrorFormatter("cidrv6", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not CIDRv6", fe.Field())
	})

	// Internet Protocol Address IP
	RegisterValidateErrorFormatter("ip_addr", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not a resolvable ip address", fe.Field())
	})

	// Internet Protocol Address IPv4
	RegisterValidateErrorFormatter("ip4_addr", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not a resolvable ipv4 address", fe.Field())
	})

	// Internet Protocol Address IPv6
	RegisterValidateErrorFormatter("ip6_addr", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not a resolvable ipv6 address", fe.Field())
	})

	// Unix domain socket end point Address
	RegisterValidateErrorFormatter("unix_addr", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not unix address", fe.Field())
	})

	// Media Access Control Address MAC
	RegisterValidateErrorFormatter("mac", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not MAC", fe.Field())
	})

	// Hostname RFC 952
	RegisterValidateErrorFormatter("hostname", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not hostname", fe.Field())
	})

	// Hostname RFC 1123
	RegisterValidateErrorFormatter("hostname_rfc1123", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not RFC 1123 hostname", fe.Field())
	})

	// Full Qualified Domain Name (FQDN)
	RegisterValidateErrorFormatter("fqdn", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not FQDN", fe.Field())
	})

	// URL Encoded
	RegisterValidateErrorFormatter("url_encoded", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not encoded url", fe.Field())
	})

	// HostPort
	RegisterValidateErrorFormatter("hostname_port", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not hostname and port", fe.Field())
	})

	// Datetime
	RegisterValidateErrorFormatter("datetime", func(fe validator.FieldError) error {
		return fmt.Errorf("the value of '%s' is not datetime", fe.Field())
	})
}
