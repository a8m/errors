package errors_test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/a8m/errors"
)

// --------------------------------------------------------
// Examples 1

type Parser struct {
	errors.Handler
}

func NewParser() *Parser {
	p := new(Parser)
	// <init code>

	// custom assert error.
	p.AssertFunc = func(msg string) error {
		return &ParseError{msg}
	}
	return p
}

func (p *Parser) Parse(b []byte) (params *Params, err error) {
	// catch only specific errors.
	defer p.Catch(&err, &json.InvalidUnmarshalError{}, &ParseError{})
	p.Must(json.Unmarshal(b, params))
	p.Expect(params.Limit > 0, "Limit must be greater than 0")
	p.Expect(params.Offset >= 0, "Offset must be greater than or equal to 0")
	// call private methods.
	p.parseDate(params)
	return
}

func (p *Parser) parseDate(params *Params) {
	// parse "created_at" field.
	v, ok := params.Filter["created_at"]
	p.Expect(ok, "created_at is a required field")
	vs, ok := v.(string)
	p.Expect(ok, "created_at must be type string")
	created, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	params.CreatedAt = created
	// parse "updated_at" field.
	v, ok = params.Filter["updated_at"]
	p.Expect(ok, "created_at is a required field")
	vs, ok = v.(string)
	p.Expect(ok, "updated_at must be type string")
	updated, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	params.UpdatedAt = updated
}

type Params struct {
	Limit     int
	Offset    int
	Filter    map[string]interface{}
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ParseError struct {
	msg string
}

func (p ParseError) Error() string { return "" }

// --------------------------------------------------------
// Example 2

// Fancy logger
type Logger struct {
	errors.Handler
	// ...
	FileName string
}

func (l *Logger) Log(v interface{}) (err error) {
	l.Catch(&err)
	buf, err := json.Marshal(v)
	l.Must(err)
	f, err := os.OpenFile(l.FileName, os.O_APPEND|os.O_WRONLY, 0644)
	l.Must(err)
	defer l.Must(f.Close())
	_, err = f.Write(buf)
	l.Must(err)
	// ...
	return
}

func TestLogger(t *testing.T) {
	l := new(Logger)
	fmt.Println(l.Log(nil))
}
