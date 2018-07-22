### errors
errors is an experimental package for errors handling in Go that simplifies the `err != nil` flow control and makes the code much readable and easier to maintain. It adopted a pattern that was taken from the Go standard libraries and made it more generic and friendly to use. You can read more about it in the section below.

#### Examples
##### Embed the `errors.Handler` in a struct.
```go
type Parser struct {
	errors.Handler
}

func (p *Parser) Parse(b []byte) (params Params, err error) {
	defer p.Catch(&err)
	p.Must(json.Unmarshal(b, &params))
	p.Assert(params.Limit > 0, &ParseError{msg: "Limit must be > 0"})
	p.Assertf(params.Offset >= 0, "Offset must be >= 0. got: %v", params.Offset)
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
	// return `ParseError` if assertion failed.
	p.Assertf(ok, "created_at is a required field")
	return
}
```
##### Using `Must` and `Assert/f` in the main function.
```go
func main() {
	// parse configuration.
	var c Config
	errors.Must(envconfig.Process("app", &c))

	// set up db connection.
	db, err := gorm.Open("mysql", c.MySQLDSN)
	errors.Must(err)

	// create elastic client.
	errors.Assertf(validator.URL(c.ElasticURL), "invalid elastic url")
	client, err := elastic.NewClient(elastic.SetURL(c.ElasticURL))
	errors.Must(err)

	// application setup.

	d, err := deleter.New(&deleter.Config{
		DB:  db,
		Log: logrus.WithField("pkg", "deleter"),
	})
	errors.Must(err)
	go d.Start()
	defer d.Stop()

	p, err := producer.New(&producer.Config{
		Queue: c.Queue,
		Log:   logrus.WithField("pkg", "producer"),
	})
	errors.Must(err)

	h := rest.NewHandler(p, d)
	log.Fatal(http.ListenAndServe(":8080", h))
}
```


### Motivation
A few years ago I wrote a Parser for a project that I was working on. Parser logic was full of deeply nested and recursive function calls, where almost every function returned an error that was bubbled up all the way to the user.
I didn't like. It was really hard to write code like this, where almost every step is an expectation. I didn't see any value in handling the errors if all I want to do is to return them to the user. I decided to give a look to the `go/parser` package in the standard library in order to learn idiomatic Go. I found this code ([1], [2]) and decided to adopt this pattern to my project.
The change was amazing. My parser was far more readable, it was easier to add or refactor code, and I just loved it like this.
Since then, almost every time I need to write a parser or anything else that similar in the complexity, I use this pattern. After too many times of copy-pasting this pattern, I decided to create this package. I guess it will help others as well.

__"You talked about parsers, but you showed above a `main` example?"__ - Yes, I treat the `main` function the same. In the sense that if I expect something to pass in order to start the application, I don't see any point in handling the error if all I want, is to crash. In these cases, I use that too.

__"Where else this pattern used in the standard packages?"__ - Like it was mentioned above, this pattern is really common in programs where almost every step is an expectation. Therefore, you can find it in packages like: [`fmt`](fmt), [`template`](template), [`template/parse`](template/parse), [`encoding/json`](json), [`encoding/gob`](gob) and more. Oh, and of course, in the `parser` package [1], [2].

__"What about performance?"__ - There is an overhead, but it's not so bad. Although, it should be improved in Go 1.11, since the compiler inlines panic calls. I will add a perf section really soon. Also for Go 1.11. Until then, you can check out #8 and #9.

__"Should I replace all my error handling with this pattern?"__ - No. There is no real rule for that, but try to find the right balance. Do not be afraid to use it, but do not abuse it.

[1]: https://github.com/golang/go/blob/ceb8fe45da7042b20189de0b66db5b33bb589f7b/src/go/parser/interface.go#L93-L98
[2]: https://github.com/golang/go/blob/ceb8fe45da7042b20189de0b66db5b33bb589f7b/src/go/parser/parser.go#L344-L364
[fmt]: https://github.com/golang/go/blob/ceb8fe45da7042b20189de0b66db5b33bb589f7b/src/fmt/scan.go#L1041-L1055
[template]: https://github.com/golang/go/blob/bedfa4e1c37bd08063865da628f242d27ca06ec4/src/text/template/exec.go#L204
[template/parse]: https://github.com/golang/go/blob/bedfa4e1c37bd08063865da628f242d27ca06ec4/src/text/template/parse/parse.go#L229
[json]: https://github.com/golang/go/blob/ceb8fe45da7042b20189de0b66db5b33bb589f7b/src/encoding/json/decode.go#L133-L140
[gob]: https://github.com/golang/go/blob/bedfa4e1c37bd08063865da628f242d27ca06ec4/src/encoding/gob/decode.go#L1086
