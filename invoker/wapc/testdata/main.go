package main

import (
	wapc "github.com/wapc/wapc-guest-tinygo"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{
		"greet": greet,
	})
}

func greet(payload []byte) ([]byte, error) {
	return []byte("hello, " + string(payload)), nil
}
