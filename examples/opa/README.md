
```bash
❯ tinygo build -o ./wasm/main.wasm -target=wasi -scheduler=none -no-debug ./wasm/main.go

❯ go run . eval 'greet("foo")' --format=raw
hello, foo
```