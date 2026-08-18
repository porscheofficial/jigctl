package hcr

import (
	"reflect"
	"testing"
)

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func mustEqual[T comparable](t *testing.T, want, got T, message ...string) {
	t.Helper()
	if want != got {
		if len(message) > 0 {
			t.Fatalf("%s\nwant: %v\ngot:  %v", message[0], want, got)
		}
		t.Fatalf("want %v, got %v", want, got)
	}
}

func mustDeepEqual(t *testing.T, want, got any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("want %#v, got %#v", want, got)
	}
}
