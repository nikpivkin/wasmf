package wapc

import (
	"context"
	"os"

	wapcgo "github.com/wapc/wapc-go"
	"github.com/wapc/wapc-go/engines/wazero"
)

type Invoker struct {
	module   wapcgo.Module
	instance wapcgo.Instance
}

func NewInvoker(ctx context.Context, src []byte) (*Invoker, error) {
	module, err := wazero.Engine().New(ctx, wapcgo.NoOpHostCallHandler, src, &wapcgo.ModuleConfig{
		Logger: wapcgo.PrintlnLogger,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
	if err != nil {
		return nil, err
	}

	instance, err := module.Instantiate(ctx)
	if err != nil {
		return nil, err
	}

	return &Invoker{
		instance: instance,
		module:   module,
	}, nil
}

func (r *Invoker) Run(ctx context.Context, fnName string, input []byte) ([]byte, error) {
	return r.instance.Invoke(ctx, fnName, input)
}

func (r *Invoker) Close(ctx context.Context) {
	r.module.Close(ctx)
	r.instance.Close(ctx)
}
