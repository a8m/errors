package errors_test

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/a8m/errors"
	"github.com/stretchr/testify/assert"
)

// --------------------------------------------------------
// Examples 1

type Parser struct {
	errors.Handler
}

func (p *Parser) Parse(b []byte) (params Params, err error) {
	defer p.Catch(&err)
	p.Must(json.Unmarshal(b, &params))
	p.Assert(params.Limit > 0, &ParseError{msg: "Limit must be greater than 0"})
	p.Assertf(params.Offset >= 0, "Offset must be greater than or equal to 0. got: %v", params.Offset)
	p.parseDate(&params)
	return
}

func (p *Parser) parseDate(params *Params) {
	// parse "created_at" field.
	v, ok := params.Filter["created_at"]
	p.Assert(ok, &ParseError{msg: "created_at is a required field"})
	vs, ok := v.(string)
	p.Assert(ok, &ParseError{msg: "created_at must be type string"})
	created, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	params.CreatedAt = created
	// parse "updated_at" field.
	v, ok = params.Filter["updated_at"]
	p.Assert(ok, &ParseError{msg: "created_at is a required field"})
	vs, ok = v.(string)
	p.Assert(ok, &ParseError{msg: "updated_at must be type string"})
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

func TestParser(t *testing.T) {
	t.Parallel()
	p := new(Parser)
	_, err := p.Parse([]byte(`{ "limit": -1 }`))
	assert.NotNil(t, err)
	assert.IsType(t, err, &ParseError{})
	_, err = p.Parse([]byte(`{ "limit": 100, "offset": -1 }`))
	assert.NotNil(t, err)
	assert.IsType(t, err, errors.AssertError{})
	p.AssertError = func(s string) error { return &ParseError{s} }
	_, err = p.Parse([]byte(`{ "limit": 100, "offset": -1 }`))
	assert.NotNil(t, err)
	assert.IsType(t, err, &ParseError{})
}

// --------------------------------------------------------
// Example 2

// Fancy logger
type Logger struct {
	errors.Handler
	FileName string
}

func (l *Logger) Log(v interface{}) (err error) {
	// catch only specific errors.
	defer l.Catch(&err, &os.PathError{})
	buf, err := json.Marshal(v)
	l.Must(err)
	f, err := os.OpenFile(l.FileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	l.Must(err)
	defer func() { l.Must(f.Close()) }()
	_, err = f.Write(buf)
	l.Must(err)
	return
}

func TestLogger(t *testing.T) {
	t.Parallel()
	const logFile = "/tmp/test.log"
	defer os.Remove(logFile)
	tests := []struct {
		path    string
		wantErr bool
	}{
		{path: "", wantErr: true},
		{path: logFile, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			l := &Logger{FileName: tt.path}
			err := l.Log(nil)
			assert.Equal(t, tt.wantErr, err != nil)
		})
	}
}

// testStruct is used to test catch functionality
type testStruct struct {
	errors.Handler
	ErrorToTest   error
	ErrorsToCache []error
}

func (ts *testStruct) Test() (err error) {
	defer ts.Catch(&err, ts.ErrorsToCache...)
	ts.Must(ts.ErrorToTest)
	return
}

func TestRuntimeError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		ts        testStruct
		wantPanic bool
		wantErr   bool
	}{
		{
			name:      "runtime error",
			ts:        testStruct{ErrorToTest: &runtime.TypeAssertionError{}},
			wantPanic: true,
		},
		{
			name:    "expected custom error with no defined errors to catch",
			ts:      testStruct{ErrorToTest: fmt.Errorf("failed")},
			wantErr: true,
		},
		{
			name: "expected custom error with defined errors to catch",
			ts: testStruct{
				ErrorToTest:   fmt.Errorf("failed"),
				ErrorsToCache: []error{fmt.Errorf("")},
			},
			wantErr: true,
		},
		{
			name: "unexpected custom error with defined errors to catch",
			ts: testStruct{
				ErrorToTest:   &ParseError{msg: "unexpected"},
				ErrorsToCache: []error{fmt.Errorf("")},
			},
			wantPanic: true,
		},
		{
			name: "no error was returned",
		},
		{
			name: "no error was returned with defined errors to catch",
			ts:   testStruct{ErrorsToCache: []error{fmt.Errorf("")}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() { tt.ts.Test() })
			} else {
				assert.Equal(t, tt.wantErr, tt.ts.Test() != nil)
			}
		})
	}
}
