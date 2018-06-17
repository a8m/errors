package errors_test

import (
	"strconv"
	"testing"

	"github.com/a8m/errors"
)

func BenchmarkMust(b *testing.B) {
	var t tester
	for i := 0; i < b.N; i++ {
		t.testMust("1") // success
		t.testMust("x") // failure
	}
}

func BenchmarkNoMust(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testNoMust("1") // success
		testNoMust("x") // failure
	}
}

type tester struct {
	errors.Handler
}

func (t *tester) testMust(s string) (err error) {
	defer t.Catch(&err)
	_, err = strconv.Atoi(s)
	t.Must(err)
	_, err = strconv.Atoi(s + s)
	t.Must(err)
	_, err = strconv.Atoi(s + s + s)
	t.Must(err)
	return
}

func testNoMust(s string) error {
	if _, err := strconv.Atoi(s); err != nil {
		return err
	}
	if _, err := strconv.Atoi(s + s); err != nil {
		return err
	}
	if _, err := strconv.Atoi(s + s + s); err != nil {
		return err
	}
	return nil
}
