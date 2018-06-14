package errors_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/a8m/x/errors"
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
	p.AssertError = func(msg string) error {
		return &ParseError{msg}
	}
	return p
}

func (p *Parser) Parse(b []byte) (params Params, err error) {
	defer p.Catch(&err)
	p.Must(json.Unmarshal(b, &params))
	p.Assert(params.Limit > 0, "Limit must be greater than 0")
	p.Assert(params.Offset >= 0, "Offset must be greater than or equal to 0")
	p.parseDate(&params)
	return
}

func (p *Parser) parseDate(params *Params) {
	// parse "created_at" field.
	v, ok := params.Filter["created_at"]
	p.Assert(ok, "created_at is a required field")
	vs, ok := v.(string)
	p.Assert(ok, "created_at must be type string")
	created, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	params.CreatedAt = created
	// parse "updated_at" field.
	v, ok = params.Filter["updated_at"]
	p.Assert(ok, "created_at is a required field")
	vs, ok = v.(string)
	p.Assert(ok, "updated_at must be type string")
	updated, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	params.UpdatedAt = updated
}

type Params struct {
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	Filter    map[string]interface{} `json:"filter,omitempty"`
	CreatedAt time.Time              `json:"created_at,omitempty"`
	UpdatedAt time.Time              `json:"updated_at,omitempty"`
}

type ParseError struct {
	msg string
}

func (e ParseError) Error() string { return e.msg }

// --------------------------------------------------------
// Example 2

// Fancy logger
type Logger struct {
	errors.Handler
	FileName string
}

func (l *Logger) Log(v interface{}) (err error) {
	// catch only specific errors.
	defer l.Catch(&err, &json.InvalidUnmarshalError{}, &ParseError{})
	buf, err := json.Marshal(v)
	l.Must(err)
	f, err := os.OpenFile(l.FileName, os.O_APPEND|os.O_WRONLY, 0644)
	l.Must(err)
	defer l.Must(f.Close())
	_, err = f.Write(buf)
	l.Must(err)
	return
}

func TestLogger(t *testing.T) {
	l := new(Logger)
	err := l.Log(nil)
	if err == nil {
		t.Fatal("expect error to not be nil")
	}
}

func TestParser(t *testing.T) {
	p := NewParser()
	_, err := p.Parse([]byte(`{ "limit": -1 }`))
	if err == nil {
		t.Fatal("expect error to not be nil")
	}
	_, ok := err.(*ParseError)
	if !ok {
		t.Fatalf("expect error to be ParseError, but got: %v", err)
	}
}
