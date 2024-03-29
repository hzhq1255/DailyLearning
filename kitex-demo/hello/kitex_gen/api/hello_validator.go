// Code generated by Validator v0.1.4. DO NOT EDIT.

package api

import (
	"bytes"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// unused protection
var (
	_ = fmt.Formatter(nil)
	_ = (*bytes.Buffer)(nil)
	_ = (*strings.Builder)(nil)
	_ = reflect.Type(nil)
	_ = (*regexp.Regexp)(nil)
	_ = time.Nanosecond
)

func (p *Request) IsValid() error {
	if len(p.Message) > int(8) {
		return fmt.Errorf("field Message max_len rule failed, current value: %d", len(p.Message))
	}
	_src := "kitex-"
	if !strings.HasPrefix(p.Message, _src) {
		return fmt.Errorf("field Message prefix rule failed, current value: %v", p.Message)
	}
	return nil
}
func (p *Response) IsValid() error {
	return nil
}
func (p *AddRequest) IsValid() error {
	return nil
}
func (p *AddResponse) IsValid() error {
	return nil
}
