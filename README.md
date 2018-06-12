### Thoughts
- Change `Expect` to `Assert`
- Add stack trace for errors
- Do we want to Give context to errors like this: `h.Must(error, string, ...interface{})`
  I'm not sure we need this if we have stack trace.
- Add package level functions
  ```go
  errors.Must(err)
  errors.Expect(expr)
  ```

### Examples

#### Simple

```go
type Logger struct {
	errors.Handler
	FileName string
}

func (l *Logger) Log(v interface{}) (err error) {
	// catch all errors except runtime.Error
	defer l.Catch(&err)
	buf, err := json.Marshal(v)
	l.Must(err)
	f, err := os.OpenFile(l.FileName, os.O_APPEND|os.O_WRONLY, 0644)
	l.Must(err)
	defer l.Must(f.Close())
	_, err = f.Write(buf)
	l.Must(err)
	return
}
```


#### Set custom assertion error
```go
type Parser struct {
    errors.Handler
    // ...
}

func NewParser() *Parser {
    p := new(Parser)
    p.Trace = true
    // make Expect throwing ParseError.
    p.AssertFunc = func(msg string) error {
    	return &ParseError{msg}
    }
    return p
}

func (p *Parser) Parse(b []byte) (params *Params, err error) {
	// catch only these errors.
	defer p.Catch(&err, &json.InvalidUnmarshalError{}, &ParseError{})
	p.Must(json.Unmarshal(b, &params))
	p.Expect(params.Limit > 0, "Limit must be greater than 0")
	p.Expect(params.Offset >= 0, "Offset must be greater than or equal to 0")
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
	p.Expect(ok, "updated_at is a required field")
	vs, ok = v.(string)
	p.Expect(ok, "updated_at must be type string")
	updated, err := time.Parse(time.RFC3339, vs)
	p.Must(err)
	params.UpdatedAt = updated
	return
}

// Let see how the function above looks like without the error handling.
func (p *Parser) Parse(b []byte) (params *Params, err error) {
    if err = json.Unmarshal(b, &params); err != nil {
    	return nil, err
    }
    if params.Limit > 0 {
    	return nil, errors.New("Limit must be greater than 0")
    }
    if params.Offset >= 0 {
    	return nil, errors.New("Offset must be greater than or equal to 0")
    }
    // parse "created_at" field.
    v, ok := params.Filter["created_at"]
    if !ok {
        return errors.New("created_at is a required field")
    }
    vs, ok := v.(string)
    if !ok {
        return errors.New("created_at must be type string")
    }
    created, err := time.Parse(time.RFC3339, vs)
    if err != nil {
        return err
    }
    params.CreatedAt = created
    // parse "updated_at" field.
    v, ok = params.Filter["updated_at"]
    if !ok {
       return errors.New("created_at is a required field")
    }
    vs, ok = v.(string)
    if !ok {
        return errors.New("updated_at must be type string")
    }
    updated, err := time.Parse(time.RFC3339, vs)
    if err != nil {
        return err
    }
    params.UpdatedAt = updated
    return nil
}
```

### Why
- Makes life eaiser when every step is an expectation or when you have deeply nested function calls.
  Writing parsers for example.
- Common in the standard library (gob, json, template, ...), and I have a few projects that use this technique.
  __Why not creating something generic__?
