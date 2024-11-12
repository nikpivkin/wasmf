package main

import (
	"bytes"

	"github.com/vmihailenco/msgpack/v5"
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{
		"greet": greet,
	})
}

func greet(payload []byte) ([]byte, error) {

	dec := msgpack.NewDecoder(bytes.NewBuffer(payload))
	dec.UsePreallocateValues(false)

	name, err := dec.DecodeString()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)

	if err := enc.EncodeString("hello, " + name); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
