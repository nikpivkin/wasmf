package main

import (
	"bytes"
	"context"
	"fmt"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/rego"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/types"
	"github.com/vmihailenco/msgpack/v5"
)

type Function struct {
	Decl *rego.Function
	Impl rego.BuiltinDyn
}

type invoker interface {
	Run(ctx context.Context, fnName string, input []byte) ([]byte, error)
}

func ToRego(cfg Config, i invoker) ([]Function, error) {
	funcs := make([]Function, 0, len(cfg.Functions))
	for _, fn := range cfg.Functions {
		if len(fn.Returns) != 1 {
			return nil, fmt.Errorf("function must return exactly 1 value, but returned %d", len(fn.Returns))
		}
		args := make([]types.Type, 0, len(fn.Parameters))
		for _, param := range fn.Parameters {
			args = append(args, toRegoType(param))
		}
		funcs = append(funcs, Function{
			Decl: &rego.Function{
				Name: fn.Name,
				Decl: types.NewFunction(args, toRegoType(fn.Returns[0])),
			},
			Impl: func(bctx rego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
				if len(terms) < len(args) {
					return nil, fmt.Errorf("mismatched argument count: expected %d, but got %d", len(terms), len(args))
				}

				var buf bytes.Buffer
				enc := msgpack.NewEncoder(&buf)
				enc.UseCompactInts(true)

				if err := marshall(enc, terms, fn.Parameters); err != nil {
					return nil, err
				}

				res, err := i.Run(bctx.Context, fn.Name, buf.Bytes())
				if err != nil {
					return nil, err
				}

				dec := msgpack.NewDecoder(bytes.NewBuffer(res))
				term, err := unmarshall(dec, fn.Returns[0])
				if err != nil {
					return nil, err
				}
				return term, nil
			},
		})
	}

	return funcs, nil
}

func toRegoType(typ Type) types.Type {
	switch typ {
	case StringType:
		return types.S
	case IntType:
		return types.N
	default:
		return types.A
	}
}

func marshall(enc *msgpack.Encoder, terms []*ast.Term, argTypes []Type) error {
	for i, typ := range argTypes {
		arg := terms[i]
		switch typ {
		case StringType:
			astv, err := builtins.StringOperand(arg.Value, arg.Location.Row)
			if err != nil {
				return fmt.Errorf("invalid parameter type: %w", err)
			}
			if err := enc.EncodeString(string(astv)); err != nil {
				return fmt.Errorf("failed to encode stirng: %w", err)
			}
		case IntType:
			astv, err := builtins.NumberOperand(arg.Value, arg.Location.Row)
			if err != nil {
				return fmt.Errorf("invalid parameter type: %w", err)
			}
			num, ok := astv.Int64()
			if !ok {
				return fmt.Errorf("failed to convert number %v to int64", astv.String())
			}
			if err := enc.EncodeInt32(int32(num)); err != nil {
				return fmt.Errorf("failed to encode int32: %w", err)
			}
		default:
			return fmt.Errorf("unsopported type: %v", typ)
		}
	}
	return nil
}

func unmarshall(dec *msgpack.Decoder, retType Type) (*ast.Term, error) {
	switch retType {
	case StringType:
		v, err := dec.DecodeString()
		if err != nil {
			return nil, fmt.Errorf("failed to decode string: %w", err)
		}
		return ast.StringTerm(v), nil
	case IntType:
		v, err := dec.DecodeInt32()
		if err != nil {
			return nil, fmt.Errorf("failed to decode string: %w", err)
		}
		return ast.IntNumberTerm(int(v)), nil
	default:
		return nil, fmt.Errorf("unsopported type: %v", retType)
	}
}
