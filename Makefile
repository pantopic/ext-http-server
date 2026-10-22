work:
	go work use sdk-go
	go work use test-go
	go work use host-wazero

wasm-go:
	@cd test-go && tinygo build -buildmode=wasi-legacy -target=wasi -opt=2 -gc=leaking -scheduler=none -o ../host-wazero/test.go.wasm
wasm-zig:
	@cd test-zig && zig build --release=small
	@cp test-zig/zig-out/bin/test-zig.wasm host-wazero/test-zig.wasm

wasm: wasm-go wasm-zig

test:
	@cd host-wazero && go test . -v -cover

bench:
	@cd host-wazero && go test -bench=. -v -run=Benchmark.*

cover:
	@mkdir -p _dist
	@cd host-wazero && go test . -coverprofile=../_dist/coverage.out -v
	@go tool cover -html=_dist/coverage.out -o _dist/coverage.html

cloc:
	@cloc . --exclude-dir=_example,_dist,internal,cmd --exclude-ext=pb.go

.PHONY: all test clean
