
```bash
❯ tinygo build -o ./wasm/main.wasm -target=wasi -scheduler=none -no-debug ./wasm/main.go

❯ go run . eval 'greet("foo")' --format=raw
hello, foo

❯ go run . eval 'parse_shell("apt update; apt install -y nginx")' --format=raw
[["apt","update"],["apt","install","-y","nginx"]]

❯ go run . eval 'parse_headers("Content-Encoding: gzip\r\nLast-Modified: Tue, 20 Aug 2013 15:45:41 GMT\r\nServer: nginx/0.8.54\r\nAge: 18884\r\nVary: Accept-Encoding\r\nContent-Type: text/html\r\nCache-Control: max-age=864000, public\r\n")' --format=raw
{"Age":["18884"],"Cache-Control":["max-age=864000, public"],"Content-Encoding":["gzip"],"Content-Type":["text/html"],"Last-Modified":["Tue, 20 Aug 2013 15:45:41 GMT"],"Server":["nginx/0.8.54"],"Vary":["Accept-Encoding"]}
```