package fakehttp

import (
	"testing"
)

func TestResponseWriter(t *testing.T) {
	rw := NewResponseWriter()
	rw.Write([]byte("foo"))
	rw.Write([]byte("bar"))

	if string(rw.Output) != "foobar" {
		t.Errorf("Expected %#v, but got %#v", "foobar", string(rw.Output))
	}
}
