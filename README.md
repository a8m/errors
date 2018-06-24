### errors
errors is an experimental package for errors handling in Go that simplifies the `err != nil` flow control and makes the code much readable and easier to maintain. It adopted a pattern that was taken from the Go standard libraries and made it more generic and friendly to use. You can read more about it in the section below.

<PHOTO WITH CODE EXAMPLE>

#### Examples
##### Embed the `errors.Handler` in a struct.
```go
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
```
`Catch` catches all errors except `runtime.Error`s by default. You can ask `Catch` to catch only specific error types.
```go
func (p *Parser) Parse(b []byte) (params Params, err error) {
  defer p.Catch(&err, &json.SyntaxError{}, &time.ParseError{})
  // ...
}
```
The default error type that `Assertf` throws is `AssertError`. You can change that by providing an error constructor as follow: 
```go
type Parser struct {
	errors.Handler
}

func NewParser() *Parser {
	p := new(Parser)
	p.AssertError = func(s string) error {
		return &ParseError{s}
	}
	return p
}

func (p *Parser) Parse(b []byte) (params Params, err error) {
	defer p.Catch(&err)
	p.Must(json.Unmarshal(b, &params))
	v, ok := params.Filter["created_at"]
	// return `ParseError` is assertion failed.
	p.Assertf(ok, "created_at is a required field")
	// ...
	return
}
```
##### Using `Must` and `Assert/f` in the main function.
```go
```


### Motivation
A few years ago while writing a parser I decided to give a look to the `go/parser` package in the standard library and found this code ([1], [2]). <What does this code do?>
I adopted that to my project and the change was amazing. From 

A few years ago I wrote a Parser for a project that I was working on. Parser logic was full of deeply nested and recursive function calls, where almost every function returned an error that was bubbled up all the way to the user. I didn't like. It was really hard to write code like this where almost every step was an expectation. So, I decided to give a look to the `go/parser` package in the standard library and found this code ([1], [2]).

### Why
- Makes life easier when every step is an Assertion or when you have deeply nested function calls.
  Writing parsers for example.
- Common in the standard library (gob, json, template, ...), and I have a few projects that use this technique.
  __Why not creating something generic__?


[1]: https://github.com/golang/go/blob/ceb8fe45da7042b20189de0b66db5b33bb589f7b/src/go/parser/interface.go#L93-L98
[2]: https://github.com/golang/go/blob/ceb8fe45da7042b20189de0b66db5b33bb589f7b/src/go/parser/parser.go#L344-L364

