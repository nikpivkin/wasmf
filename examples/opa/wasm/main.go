package main

import (
	"bufio"
	"bytes"
	"net/textproto"
	"strings"

	"github.com/vmihailenco/msgpack/v5"
	wapc "github.com/wapc/wapc-guest-tinygo"
	"mvdan.cc/sh/v3/syntax"
)

func main() {
	wapc.RegisterFunctions(wapc.Functions{
		"greet":         greet,
		"parse_shell":   parseShell,
		"parse_headers": parseHeaders,
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

	if err := enc.EncodeString("hello, " + name); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func parseShell(payload []byte) ([]byte, error) {
	dec := msgpack.NewDecoder(bytes.NewBuffer(payload))
	dec.UsePreallocateValues(false)
	programm, err := dec.DecodeString()
	if err != nil {
		return nil, err
	}

	cmds, err := _parseShell(programm)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err := enc.EncodeArrayLen(len(cmds)); err != nil {
		return nil, err
	}
	for _, cmd := range cmds {
		if err := enc.EncodeArrayLen(len(cmd)); err != nil {
			return nil, err
		}
		for _, tok := range cmd {
			if err := enc.EncodeString(tok); err != nil {
				return nil, err
			}
		}
	}

	return buf.Bytes(), nil
}

func _parseShell(programm string) ([][]string, error) {
	f, err := syntax.NewParser().Parse(strings.NewReader(programm), "")
	if err != nil {
		return nil, err
	}

	printer := syntax.NewPrinter()

	var commands [][]string
	syntax.Walk(f, func(node syntax.Node) bool {
		switch x := node.(type) {
		case *syntax.CallExpr:
			args := x.Args
			var cmd []string
			for _, word := range args {
				var buffer bytes.Buffer
				printer.Print(&buffer, word)
				cmd = append(cmd, buffer.String())
			}
			if cmd != nil {
				commands = append(commands, cmd)
			}
		}
		return true
	})

	return commands, nil
}

func parseHeaders(payload []byte) ([]byte, error) {
	dec := msgpack.NewDecoder(bytes.NewBuffer(payload))
	dec.UsePreallocateValues(false)
	raw, err := dec.DecodeString()
	if err != nil {
		return nil, err
	}
	headers, err := _parseHeaders(raw)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	if err := enc.EncodeMapLen(len(headers)); err != nil {
		return nil, err
	}
	for k, v := range headers {
		if err := enc.EncodeString(k); err != nil {
			return nil, err
		}
		if err := enc.EncodeArrayLen(len(v)); err != nil {
			return nil, err
		}
		for _, vv := range v {
			if err := enc.EncodeString(vv); err != nil {
				return nil, err
			}
		}
	}
	return buf.Bytes(), nil
}

func _parseHeaders(raw string) (map[string][]string, error) {
	reader := bufio.NewReader(strings.NewReader(raw + "\r\n"))
	tp := textproto.NewReader(reader)

	mimeHeader, err := tp.ReadMIMEHeader()
	if err != nil {
		return nil, err
	}

	return mimeHeader, nil
}
