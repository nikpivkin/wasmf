package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/cmd"
	oparego "github.com/open-policy-agent/opa/rego"

	"github.com/nikpivkin/wasmf/invoker/wapc"
)

func main() {
	flag.Parse()

	ctx := context.Background()

	wasm, err := os.ReadFile("wasm/main.wasm")
	if err != nil {
		panic(err)
	}

	invoker, err := wapc.NewInvoker(ctx, wasm)
	if err != nil {
		panic(err)
	}

	cfg := Config{
		Functions: []FunctionDeclaration{
			{
				Name:       "greet",
				Parameters: []*Type{{Kind: StringType}},
				Returns:    []*Type{{Kind: StringType}},
			},
			{
				Name:       "parse_shell",
				Parameters: []*Type{{Kind: StringType}},
				Returns: []*Type{
					{
						Kind:      ArrayType,
						ValueType: &Type{Kind: ArrayType, ValueType: &Type{Kind: StringType}},
					},
				},
			},
		},
	}

	funcs, err := ToRego(cfg, invoker)
	if err != nil {
		panic(err)
	}

	for _, fn := range funcs {
		oparego.RegisterBuiltinDyn(fn.Decl, func(bctx oparego.BuiltinContext, terms []*ast.Term) (*ast.Term, error) {
			ret, err := fn.Impl(bctx, terms)
			if err != nil {
				log.Printf("Failed to call function %q: %s", fn.Decl.Name, err)
				return nil, err
			}
			return ret, err
		})
	}

	if err := cmd.RootCommand.Execute(); err != nil {
		panic(err)
	}
}
