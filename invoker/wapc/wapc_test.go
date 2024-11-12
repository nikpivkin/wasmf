package wapc_test

import (
	"context"
	"os"
	"testing"

	"github.com/nikpivkin/wasmf/invoker/wapc"
)

func TestInvoke(t *testing.T) {
	b, err := os.ReadFile("testdata/main.wasm")
	if err != nil {
		t.Fatal(err)
	}

	r, err := wapc.NewInvoker(context.TODO(), b)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close(context.TODO())
	res, err := r.Run(context.TODO(), "greet", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}

	expected := "hello, foo"
	if string(res) != expected {
		t.Fatalf("expected %q, but got %s", expected, string(res))
	}
}
