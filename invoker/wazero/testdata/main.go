package main

// #include <stdlib.h>
import "C"

import (
	"bytes"
	"unsafe"

	"github.com/vmihailenco/msgpack/v5"
)

func _add(a, b uint32) uint32 {
	return a + b
}

//export add
func add(ptr, size uint32) uint64 {
	data := ptrToBytes(ptr, size)
	dec := msgpack.NewDecoder(bytes.NewBuffer(data))
	dec.UsePreallocateValues(false)

	a, _ := dec.DecodeInt32()
	b, _ := dec.DecodeInt32()
	ret := _add(uint32(a), uint32(b))

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	enc.UseCompactInts(true)

	if err := enc.EncodeInt32(int32(ret)); err != nil {
		return 0
	}

	ptr, size = bytesToPtr(buf.Bytes())
	return (uint64(ptr) << uint64(32)) | uint64(size)
}

func _greet(s string) string {
	return "hello, " + s
}

//export greet
func greet(ptr, size uint32) uint64 {
	s := ptrToString(ptr, size)
	ret := _greet(string(s))
	ptr, size = stringToPtr(ret)
	return (uint64(ptr) << uint64(32)) | uint64(size)
}

// ptrToString returns a string from WebAssembly compatible numeric types
// representing its pointer and length.
func ptrToString(ptr uint32, size uint32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

// ptrToString returns a string from WebAssembly compatible numeric types
// representing its pointer and length.
func ptrToBytes(ptr uint32, size uint32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}

func stringToPtr(s string) (uint32, uint32) {
	size := C.ulong(len(s))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), s)
	return uint32(uintptr(ptr)), uint32(size)
}

func bytesToPtr(b []byte) (uint32, uint32) {
	size := C.ulong(len(b))
	ptr := unsafe.Pointer(C.malloc(size))
	copy(unsafe.Slice((*byte)(ptr), size), b)
	return uint32(uintptr(ptr)), uint32(size)
}

func main() {}
