package main

/*
The go:generate directive runs bpf2go.

bpf2go:
1. Uses Clang to compile exec.bpf.c into eBPF bytecode.
2. Generates Go bindings for the compiled eBPF program.
3. Embeds the compiled bytecode into generated Go code.

"exec" is the prefix used for generated Go types.

The -I flag tells Clang where to find vmlinux.h.

This directive runs when we execute:
    go generate ./cmd/kernelscope

bpf2go does NOT compile Go code into eBPF.
*/

//go:generate go tool bpf2go -tags linux exec ../../bpf/exec.bpf.c -- -I../../bpf
