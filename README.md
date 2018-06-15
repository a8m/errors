### Thoughts
- use github.com/pkg/errors to get the stack trace.

### Examples

```go
type Parser struct {
    errors.Handler
}

func (p *Parser) Parse(b []byte) (params *Params, err error) {
	defer p.Catch(&err)
	p.Must(json.Unmarshal(b, &params))
	v, ok := params.Filter["created_at"]
	p.Assertf(ok, "created_at is a required field")
	vs, ok := v.(string)
	p.Assertf(ok, "created_at must be type string")
	created, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	return
}
```
__Let see how the function above looks like without the error handling__
```go
func (p *Parser) Parse(b []byte) (params *Params, err error) {
    if err = json.Unmarshal(b, &params); err != nil {
    	return nil, err
    }
    v, ok := params.Filter["created_at"]
    if !ok {
        return nil, errors.New("created_at is a required field")
    }
    vs, ok := v.(string)
    if !ok {
        return nil, errors.New("created_at must be type string")
    }
    created, err := time.Parse(time.RFC3339, vs)
    if err != nil {
        return nil, err
    }
    return params, nil
}
```

# Example 2
```go
type Logger struct {
	errors.Handler
	FileName string
}

func (l *Logger) Log(v interface{}) (err error) {
	// catch only these errors.
	defer l.Catch(&err, &json.InvalidUnmarshalError{}, &ParseError{})
	buf, err := json.Marshal(v)
	l.Must(err)
	f, err := os.OpenFile(l.FileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	l.Must(err)
	defer l.Must(f.Close())
	_, err = f.Write(buf)
	l.Must(err)
	return
}
```

### Why
- Makes life easier when every step is an Assertion or when you have deeply nested function calls.
  Writing parsers for example.
- Common in the standard library (gob, json, template, ...), and I have a few projects that use this technique.
  __Why not creating something generic__?
