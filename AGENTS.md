# AGENTS.md — KernelScope

Instructions for AI coding agents working on KernelScope.

## Project mission

KernelScope is a Linux-first, Go-based observability tool powered by eBPF. It should help a developer understand what is happening on a Linux system by tracing process execution, file activity, network connections, and performance events.

This is also a learning project. The primary maintainer is learning eBPF by building KernelScope and uses *Learning eBPF* by Liz Rice as a theory reference. Prioritize understanding, small experiments, and explainable code over feature velocity.

## Scope and priorities

- Current milestone: a CLI process-execution tracer using the `sched/sched_process_exec` tracepoint.
- Initial user experience: run `sudo ./kernelscope watch exec` and see timestamped process events. The exact CLI syntax may evolve; keep it simple and document changes.
- Later milestones, in order: file activity, TCP connection visibility, syscall latency/performance, interactive terminal UI, incident timeline.
- Not in scope yet: Kubernetes, containers/orchestration integrations, distributed agents, cloud backends, databases, elaborate frontend frameworks, or production deployment.
- Start with observation, not enforcement: no packet dropping, process blocking, or security policy enforcement unless explicitly requested.

## Development environment

- Development target: Rocky Linux 10.2 inside a Vagrant VM, preferably with KVM/libvirt.
- Source is edited on the host and synced into the VM; follow the repository's `Vagrantfile` configuration instead of assuming live sync.
- Run eBPF compilation, loading, and kernel tests inside the Rocky Linux VM, not on the host by default.
- Main application language: Go. Kernel-side eBPF programs: small C files compiled by Clang/LLVM.
- Preferred Go library: `github.com/cilium/ebpf`, with `bpf2go` for compiling programs and generating Go bindings.
- Prefer kernel BTF and CO-RE where useful; check `/sys/kernel/btf/vmlinux` before relying on it.
- Confirm installed tool versions and actual kernel capabilities before adding dependencies or assuming a feature is available.

## Architecture

```text
Linux kernel event (tracepoint / probe)
    -> small C eBPF program
    -> eBPF map or ring buffer
    -> Go agent (cilium/ebpf)
    -> CLI output (TUI later)
```

The Go process loads, attaches, reads from, and detaches eBPF programs. The C program executes in kernel context only when its hook fires. Do not put general application logic in the kernel.

For initial learning experiments, `bpf_printk` and `trace_pipe` are acceptable. For KernelScope features, prefer structured events delivered through a ring buffer; do not treat tracing logs as a production event API.

## Repository conventions

Follow the existing repository layout. Suggested locations if not yet present:

```text
kernelscope/
├── Vagrantfile
├── Makefile
├── go.mod
├── bpf/
│   ├── exec.bpf.c
│   └── headers/             # generated/local BTF headers if needed
├── cmd/kernelscope/         # CLI entry point
├── internal/
│   ├── tracer/              # loader, attachment, reader, bpf2go directive
│   └── events/              # Go event types and decoding
├── experiments/             # isolated exercises from Learning eBPF
├── scripts/                 # VM setup / BTF generation helpers
└── docs/                    # architecture and learning notes
```

Do not create placeholder files or directories for future milestones. Keep generated `bpf2go` files next to their owning Go package, never edit generated output manually, and document how to regenerate it. Avoid committing machine-specific build artifacts or a large generated `vmlinux.h` unless the repository explicitly adopts that policy.

## How an agent should work

1. **Inspect before editing.** Read the relevant source, `go.mod`, build instructions, and Vagrant setup. Do not assume the suggested tree is already implemented.
2. **Explain the concept first.** Briefly describe the kernel hook, eBPF helper/map, or Go API involved in the requested change, in beginner-friendly language.
3. **Make one small change at a time.** Prefer the smallest working vertical slice over broad refactoring or adding many features at once.
4. **Connect theory to practice.** When useful, identify the topic to read in *Learning eBPF*; don't invent quotations or page numbers.
5. **Show the complete data path.** For tracing features, explain where the event originates, what runs in kernel space, how data reaches Go, and how it is rendered.
6. **Make verification reproducible.** Provide exact build, test, run, and manual event-generation commands, indicating whether each runs on the host or in the VM.
7. **Report limitations.** State what was actually executed and what could not be verified, including missing privileges, dependencies, or kernel support.
8. **Preserve the learning experience.** Favor understandable code and focused explanations over complex abstractions. If asked to teach, use questions or checkpoints instead of immediately introducing the entire next milestone.

## Go standards

- Write idiomatic, readable Go; use `gofmt` and check `go vet` when available.
- Keep `main` thin. Encapsulate eBPF loading, attachment, event reading, and shutdown in internal packages.
- Return useful errors with context using `%w`; do not silently ignore important failures.
- Close links, maps, ring-buffer readers, and other resources. Handle `SIGINT`/`SIGTERM` cleanly.
- Use `context.Context` for controllable long-running operations where appropriate.
- Add tests for event decoding, formatting, filtering, and other pure Go logic without requiring root or a running eBPF program.
- Do not add frameworks, databases, or dependencies without a clear need.

## eBPF/C standards

- Keep programs short, event-driven, and compatible with the verifier.
- Prefer stable tracepoints for initial tracing; justify kprobes or other less stable hooks when needed.
- Use permitted eBPF helpers and bounded memory access. Never assume ordinary C library functions are available inside eBPF programs.
- Keep a documented, fixed event structure shared between C and Go; verify field sizes, alignment, and encoding when decoding.
- Distinguish process ID (TGID), thread ID (TID), and command name (`comm`); do not call `comm` a full executable path.
- Avoid expensive work per event. Filter in kernel space only when it reduces overhead without complicating the learning objective.
- When adding a ring buffer, handle reservation/output failures and explain that events can be lost under load. Expose drop counts when practical.
- Do not attach invasive or high-overhead probes globally without considering performance and operational safety.

## Security and operational safety

- Treat privileged commands as VM-only by default. Never change the host kernel, host networking, system services, or security configuration without explicit approval.
- Do not disable SELinux, signature checks, or kernel security protections as a shortcut.
- Prefer read-only observation. Never make covert data collection, credential capture, or process tampering part of KernelScope.
- Keep sample output free of sensitive contents (arguments, paths, or payloads can expose secrets). Explain the privacy implications of collected events.
- Use a disposable VM for experiments that can destabilize the system. Separate demonstrations from production-readiness claims.

## Validation checklist for changes

- [ ] The change addresses only the requested milestone.
- [ ] New Go code is formatted and relevant unit tests pass (`go test ./...`).
- [ ] Any changed eBPF code compiles via the project's documented `bpf2go`/Clang workflow.
- [ ] Loading and attaching are tested inside the Rocky Linux VM when possible.
- [ ] A manual command reproduces the event (for example, executing `ls` or `curl`).
- [ ] All resources detach/close when KernelScope exits.
- [ ] Errors and dropped events are visible or documented.
- [ ] Documentation explains what was learned and any environment requirements.

## Definition of done: first milestone

KernelScope can attach a Go-managed eBPF program to `sched/sched_process_exec` in the Rocky Linux VM, observe newly executed commands, decode structured records in Go, print their timestamps and process identifiers, and detach cleanly on exit. A small experiment using `bpf_printk` may precede the structured-event version.

## Guiding rule

Learn the kernel concept, build the smallest working experiment, then integrate it into KernelScope. Keep Kubernetes out of the project until the maintainer explicitly requests it.
