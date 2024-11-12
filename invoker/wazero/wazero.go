package wazero

import (
	"context"
	"fmt"
	"log"

	wazerogo "github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Invoker struct {
	ctx     context.Context
	runtime wazerogo.Runtime
	module  api.Module
	malloc  api.Function
	free    api.Function
}

func NewInvoker(ctx context.Context, wasm []byte) (*Invoker, error) {
	runtime := wazerogo.NewRuntime(ctx)
	if _, err := wasi_snapshot_preview1.Instantiate(ctx, runtime); err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	module, err := runtime.Instantiate(ctx, wasm)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate module: %w", err)
	}

	malloc := module.ExportedFunction("malloc")
	free := module.ExportedFunction("free")

	if malloc == nil || free == nil {
		return nil, fmt.Errorf("malloc or free function not found in module")
	}

	return &Invoker{
		ctx:     ctx,
		runtime: runtime,
		module:  module,
		malloc:  malloc,
		free:    free,
	}, nil
}

func (l *Invoker) Run(ctx context.Context, fnName string, input []byte) ([]byte, error) {
	fnInstance := l.module.ExportedFunction(fnName)
	if fnInstance == nil {
		return nil, fmt.Errorf("function %s not found in wasm module", fnName)
	}

	argSize := uint64(len(input))
	arg, err := l.malloc.Call(ctx, argSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory in WebAssembly module with malloc: %w", err)
	}

	if len(arg) != 1 {
		return nil, fmt.Errorf("malloc returned unexpected number of values: expected 1, but got %d", len(arg))
	}

	argPtr := arg[0]
	if argPtr == 0 {
		return nil, fmt.Errorf("malloc returned a null pointer")
	}

	defer func() {
		if _, err := l.free.Call(ctx, argPtr); err != nil {
			log.Println("Error freeing memory:", err)
		}
	}()

	if !l.module.Memory().Write(uint32(argPtr), input) {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of memory bounds (memory size: %d)",
			argPtr, argSize, l.module.Memory().Size())
	}

	results, err := fnInstance.Call(ctx, argPtr, argSize)
	if err != nil {
		return nil, err
	}

	if len(results) != 1 {
		return nil, fmt.Errorf("function returned %d values, but 1 was expected", len(results))
	}

	res := results[0]

	resPtr := uint32(res >> 32)
	resSize := uint32(res)

	if resPtr == 0 {
		return nil, fmt.Errorf("function returned a null pointer; expected a valid pointer to a byte")
	}

	defer func() {
		if _, err := l.free.Call(ctx, uint64(resPtr)); err != nil {
			log.Printf("Failed to free memory at %d: %v", resPtr, err)
		}
	}()

	bytes, ok := l.module.Memory().Read(resPtr, resSize)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of memory bounds (memory size: %d)",
			resPtr, resSize, l.module.Memory().Size())
	}

	return bytes, nil
}

func (l *Invoker) Close(ctx context.Context) {
	l.runtime.Close(ctx)
	l.module.Close(ctx)
}
