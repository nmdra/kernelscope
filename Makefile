APP := kernelscope
BIN_DIR := bin
BIN := $(BIN_DIR)/$(APP)

BPF_OBJECT := cmd/kernelscope/exec_bpfel.o
BPF_PROGRAM := trace_exec

.DEFAULT_GOAL := help

.PHONY: help check generate build run \
        trace inspect programs features \
        fmt vet test clean all

# Display available commands
help:
	@echo "KernelScope - Available Commands"
	@echo ""
	@echo "  make check      Check development tools"
	@echo "  make generate   Compile eBPF and generate Go bindings"
	@echo "  make build      Build KernelScope Go binary"
	@echo "  make run        Run KernelScope with root privileges"
	@echo "  make trace      Watch kernel tracing output"
	@echo "  make inspect    Inspect compiled eBPF bytecode"
	@echo "  make programs   List loaded KernelScope eBPF programs"
	@echo "  make features   Show kernel eBPF capabilities"
	@echo "  make fmt        Format Go code"
	@echo "  make vet        Check Go code"
	@echo "  make test       Run Go tests"
	@echo "  make clean      Remove built binary"
	@echo "  make all        Generate and build"

# Check required development tools
check:
	@go version
	@clang --version | head -n 1
	@bpftool version
	@test -f /sys/kernel/btf/vmlinux && \
		echo "Kernel BTF: Available"
	@test -f /sys/kernel/tracing/events/sched/sched_process_exec/id && \
		echo "Process exec tracepoint: Available"

# Compile C into eBPF and generate Go bindings
generate:
	go generate ./cmd/kernelscope

# Build the Go application
build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN) ./cmd/kernelscope

# Run KernelScope (requires root privileges)
run: build
	sudo ./$(BIN)

# Read eBPF debug messages from trace_pipe
# Run in a second terminal
trace:
	sudo cat /sys/kernel/tracing/trace_pipe

# Disassemble the generated eBPF object
inspect:
	llvm-objdump -d $(BPF_OBJECT)

# Inspect loaded KernelScope eBPF programs
programs:
	sudo bpftool prog show name $(BPF_PROGRAM)

# Display supported kernel eBPF features
features:
	sudo bpftool feature probe kernel

# Format Go source files
fmt:
	go fmt ./...

# Run static analysis
vet:
	go vet ./...

# Run Go tests
test:
	go test ./...

# Remove build artifacts (keep generated bindings)
clean:
	rm -rf $(BIN_DIR)

# Generate eBPF and build KernelScope
all: generate build
