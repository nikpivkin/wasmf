package wazero_test

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/nikpivkin/wasmf/invoker/wazero"
)

func TestXxx1(t *testing.T) {
	b, err := os.ReadFile("testdata/main.wasm")
	if err != nil {
		t.Fatal(err)
	}

	l, err := wazero.NewInvoker(context.TODO(), b)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close(context.TODO())

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)

	enc.EncodeInt32(1)
	enc.EncodeInt32(2)

	ret, err := l.Run(context.TODO(), "add", buf.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	dec := msgpack.NewDecoder(bytes.NewBuffer(ret))
	num, err := dec.DecodeInt32()
	if err != nil {
		t.Fatal(err)
	}

	expected := int32(3)
	if num != expected {
		t.Fatalf("expected %d, but got %d", expected, num)
	}
}

func TestXxx2(t *testing.T) {
	b, err := os.ReadFile("testdata/main.wasm")
	if err != nil {
		t.Fatal(err)
	}

	l, err := wazero.NewInvoker(context.TODO(), b)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close(context.TODO())

	ret, err := l.Run(context.TODO(), "greet", []byte("foo"))
	if err != nil {
		t.Fatal(err)
	}

	expected := "hello, foo"
	if string(ret) != expected {
		t.Fatalf("expected %q, but got %s", expected, string(ret))
	}
}
