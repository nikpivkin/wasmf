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

				if err := marshal(enc, terms, fn.Parameters); err != nil {
					return nil, err
				}

				res, err := i.Run(bctx.Context, fn.Name, buf.Bytes())
				if err != nil {
					return nil, err
				}

				dec := msgpack.NewDecoder(bytes.NewBuffer(res))
				term, err := unmarshal(dec, fn.Returns[0])
				if err != nil {
					return nil, err
				}
				return term, nil
			},
		})
	}

	return funcs, nil
}

func toRegoType(typ *Type) types.Type {
	switch typ.Kind {
	case StringType:
		return types.S
	case IntType:
		return types.N
	case ArrayType:
		return types.NewArray(nil, toRegoType(typ.ValueType))
	case ObjectType:
		return types.NewObject(nil, types.NewDynamicProperty(toRegoType(typ.KeyType), toRegoType(typ.ValueType)))
	}

	panic("unsopported type: " + typ.String())
}

func marshal(enc *msgpack.Encoder, terms []*ast.Term, argTypes []*Type) error {
	for i, typ := range argTypes {
		arg := terms[i]
		if err := marshalTerm(enc, arg, typ); err != nil {
			return err
		}
	}
	return nil
}

func marshalTerm(enc *msgpack.Encoder, term *ast.Term, typ *Type) error {
	switch typ.Kind {
	case StringType:
		astv, err := builtins.StringOperand(term.Value, term.Location.Row)
		if err != nil {
			return fmt.Errorf("invalid parameter type: %w", err)
		}
		if err := enc.EncodeString(string(astv)); err != nil {
			return fmt.Errorf("failed to encode stirng: %w", err)
		}
	case IntType:
		astv, err := builtins.NumberOperand(term.Value, term.Location.Row)
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
	case ArrayType:
		astv, err := builtins.ArrayOperand(term.Value, term.Location.Row)
		if err != nil {
			return fmt.Errorf("invalid parameter type: %w", err)
		}
		if err := enc.EncodeArrayLen(astv.Len()); err != nil {
			return err
		}
		for i := range astv.Len() {
			if err := marshalTerm(enc, astv.Elem(i), typ.ValueType); err != nil {
				return fmt.Errorf("failed to marshall array element %d: %w", i, err)
			}

		}
	case ObjectType:
		astv, err := builtins.ObjectOperand(term.Value, term.Location.Row)
		if err != nil {
			return fmt.Errorf("invalid parameter type: %w", err)
		}
		if err := enc.EncodeMapLen(astv.Len()); err != nil {
			return err
		}
		if err := astv.Iter(func(t1, t2 *ast.Term) error {
			if err := marshalTerm(enc, t1, typ.KeyType); err != nil {
				return fmt.Errorf("failed to marshall key: %w", err)
			}

			if err := marshalTerm(enc, t2, typ.ValueType); err != nil {
				return fmt.Errorf("failed to marshall key: %w", err)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("failed to marshall object property: %w", err)
		}
	default:
		return fmt.Errorf("unsopported type: %v", typ.String())
	}

	return nil
}

func unmarshal(dec *msgpack.Decoder, retType *Type) (*ast.Term, error) {
	switch retType.Kind {
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
	case ArrayType:
		arraySize, err := dec.DecodeArrayLen()
		if err != nil {
			return nil, fmt.Errorf("failed to decode array len: %w", err)
		}
		elems := make([]*ast.Term, arraySize)
		for i := range arraySize {
			elemVal, err := unmarshal(dec, retType.ValueType)
			if err != nil {
				return nil, fmt.Errorf("failed to decode array element: %w", err)
			}
			elems[i] = elemVal
		}
		return ast.ArrayTerm(elems...), nil
	case ObjectType:
		objSize, err := dec.DecodeMapLen()
		if err != nil {
			return nil, fmt.Errorf("failed to decode object len: %w", err)
		}
		properties := make([][2]*ast.Term, objSize)
		for i := range objSize {
			propKey, err := unmarshal(dec, retType.KeyType)
			if err != nil {
				return nil, fmt.Errorf("failed to decode property key: %w", err)
			}
			propVal, err := unmarshal(dec, retType.ValueType)
			if err != nil {
				return nil, fmt.Errorf("failed to decode property value: %w", err)
			}
			properties[i] = [2]*ast.Term{propKey, propVal}
		}
		return ast.ObjectTerm(properties...), nil
	default:
		return nil, fmt.Errorf("unsopported type: %v", retType.String())
	}
}

// func unmarshalTerm(dec *msgpack.Decoder, retType *Type) (*ast.Term, error) {

// }
