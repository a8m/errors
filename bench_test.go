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

func BenchmarkAssert(b *testing.B) {
	var t tester
	for i := 0; i < b.N; i++ {
		t.testAssert(map[string]interface{}{"a": 1}) // success
		t.testAssert(map[string]interface{}{})       // failure
	}
}

func BenchmarkAssertf(b *testing.B) {
	var t tester
	for i := 0; i < b.N; i++ {
		t.testAssertf(map[string]interface{}{"a": 1}) // success
		t.testAssertf(map[string]interface{}{})       // failure
	}
}

func BenchmarkNoAssert(b *testing.B) {
	for i := 0; i < b.N; i++ {
		testNoAssert(map[string]interface{}{"a": 1}) // success
		testNoAssert(map[string]interface{}{})       // failure
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

func (t *tester) testAssert(v interface{}) (err error) {
	defer t.Catch(&err)
	m, ok := v.(map[string]interface{})
	t.Assert(ok, assertError{"expect map as an argument"})
	k, ok := m["a"]
	t.Assert(ok, assertError{"expect key 'a' in the map"})
	_, ok = k.(int)
	t.Assert(ok, assertError{"expect 'a' to be int"})
	return
}

func (t *tester) testAssertf(v interface{}) (err error) {
	defer t.Catch(&err)
	m, ok := v.(map[string]interface{})
	t.Assertf(ok, "expect map as an argument")
	k, ok := m["a"]
	t.Assertf(ok, "expect key 'a' in the map")
	_, ok = k.(int)
	t.Assertf(ok, "expect 'a' to be int")
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

func testNoAssert(v interface{}) (err error) {
	m, ok := v.(map[string]interface{})
	if !ok {
		return assertError{"expect map as an argument"}
	}
	k, ok := m["a"]
	if !ok {
		return assertError{"expect key 'a' in the map"}
	}
	_, ok = k.(int)
	if !ok {
		return assertError{"expect 'a' to be int"}
	}
	return
}

type assertError struct {
	string
}

func (k assertError) Error() string { return k.string }
