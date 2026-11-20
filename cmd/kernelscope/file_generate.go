package main

// Compile file.bpf.c into eBPF bytecode.
//
// The -type option additionally generates a Go type
// matching our C struct file_event.
//
//go:generate go tool bpf2go -type file_event file ../../bpf/file.bpf.c -- -I../../bpf
