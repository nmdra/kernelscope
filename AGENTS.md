# AGENTS.md — KernelScope

Instructions for AI coding agents working on KernelScope.

## Project mission

KernelScope is a Linux-first, Go-based observability tool powered by eBPF. It should help a developer understand what is happening on a Linux system by tracing process execution, file activity, network connections, and performance events.

Over time, KernelScope may grow into a **Go-based host observability agent** that exposes Prometheus metrics, sends selected alerts to Slack, offers a read-only MCP interface, and exchanges event summaries and decisions with a **System Decision Model** for classification and anomaly detection. These are future capabilities, not requirements for the current tracer.

This is also a learning project. The primary maintainer is learning eBPF by building KernelScope and uses _Learning eBPF_ by Liz Rice as a theory reference. Prioritize understanding, small experiments, and explainable code over feature velocity.

## Scope and priorities

- **Current milestone:** a CLI process-execution tracer using the `sched/sched_process_exec` tracepoint.
- **Initial user experience:** run `sudo ./kernelscope watch exec` and see timestamped process events. The exact CLI syntax may evolve; keep it simple and document changes.
- **Core observability milestones, in order:** file activity, TCP connection visibility, syscall latency/performance, interactive terminal UI, incident timeline.
- **Later integrations, only after an explicit request and a working observability foundation:** Prometheus metrics, configurable Slack alerts, the System Decision Model, and MCP support. Their implementation order is not fixed.
- **Not in scope yet:** Kubernetes, containers/orchestration integrations, distributed agents, cloud backends, databases, elaborate frontend frameworks, or production deployment.
- **Observation before enforcement:** no packet dropping, process blocking, automated remediation, or security policy enforcement unless explicitly requested.

Do not implement future integrations merely because they appear in the target architecture. A milestone should be a small, independently understandable vertical slice.

## Development environment

- Development target: Rocky Linux 10.2 inside a Vagrant VM, preferably with KVM/libvirt.
- Source is edited on the host and synced into the VM; follow the repository's `Vagrantfile` configuration instead of assuming live sync.
- Run eBPF compilation, loading, and kernel tests inside the Rocky Linux VM, not on the host by default.
- Main application language: Go. Kernel-side eBPF programs: small C files compiled by Clang/LLVM.
- Preferred Go library: `github.com/cilium/ebpf`, with `bpf2go` for compiling programs and generating Go bindings.
- Prefer kernel BTF and CO-RE where useful; check `/sys/kernel/btf/vmlinux` before relying on it.
- Confirm installed tool versions and actual kernel capabilities before adding dependencies or assuming a feature is available.

## Architecture

### Current architecture: first working tracer

```text
Linux kernel event (sched/sched_process_exec)
    -> small C eBPF tracepoint program (kernel space)
    -> eBPF ring buffer (kernel-to-user transport)
    -> Go agent / tracer (cilium/ebpf, user space)
    -> CLI output
```

The Go process loads, attaches, reads from, and detaches eBPF programs. The C program executes in kernel context only when its hook fires. Do not put general application logic in the kernel.

For initial learning experiments, `bpf_printk` and `trace_pipe` are acceptable. For KernelScope features, prefer structured events delivered through a ring buffer; do not treat tracing logs as a production event API.

### Target architecture: future, not yet implemented

```text
                        USER SPACE

                   +----------------------+
                   |       Go Agent       |
                   |  event processing,   |
                   |  routing, lifecycle  |
                   +----------------------+
                     |     |      |      |
                     |     |      |      +<--> MCP interface (read-only first)
                     |     |      +----------> Slack alerts (opt-in)
                     |     +----------------> /metrics <-- Prometheus scrapes
                     +<--------------------> System Decision Model
                       |                       classification, anomaly
                       +--> CLI / TUI          detection, recommendations

----------------------- kernel / user boundary -----------------------
                        KERNEL SPACE

Linux kernel events -> eBPF hooks -> ring buffer -> Go Agent (above)
```

This is a conceptual diagram, not an implementation checklist. The **Go Agent** is the central user-space component, similar in _role_ to a Cilium agent, but KernelScope must not adopt Cilium's complexity or internal architecture by default.

Design boundaries for future work:

- **eBPF programs:** observe kernel events and emit small, well-defined records. No model inference, alert formatting, network calls, or business logic in eBPF.
- **Go Agent:** own eBPF lifecycle, event decoding, local event processing, CLI/TUI output, and future integration adapters. Prefer simple, testable interfaces over an elaborate event framework.
- **System Decision Model:** receive selected, appropriately minimized events or summaries from Go; classify activity, identify anomalies, and return structured _recommendations/decisions_ with an explanation. The Go Agent remains responsible for checking configuration and deciding whether to notify. The model is not named “Jev”; its algorithm, runtime, and provider are **not selected**.
- **Prometheus:** scrape an HTTP metrics endpoint exposed by the Go Agent. Do not design the agent to push raw event streams into Prometheus.
- **Slack:** receive explicit, configured alert notifications from the Go Agent. Avoid alerting on every raw kernel event.
- **MCP (Model Context Protocol):** a future, deliberately scoped way for compatible external agents to inspect KernelScope observations. Start with read-only resources/tools; do not provide arbitrary shell execution or kernel modification capabilities.

Do not claim that a model is accurate, an anomaly is malicious, or an alert is a confirmed incident without supporting evidence. The model's output is advisory; **no autonomous enforcement** is part of the plan.

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

If and when future integrations are approved, add small packages such as `internal/metrics`, `internal/alerts`, `internal/decision`, or `internal/mcp` **only when implementation code needs them**. These names are suggestions, not a mandate.

Do not create placeholder files or directories for future milestones. Keep generated `bpf2go` files next to their owning Go package, never edit generated output manually, and document how to regenerate it. Avoid committing machine-specific build artifacts or a large generated `vmlinux.h` unless the repository explicitly adopts that policy.

## How an agent should work

1. **Inspect before editing.** Read the relevant source, `go.mod`, build instructions, and Vagrant setup. Do not assume the suggested tree is already implemented.
2. **Explain the concept first.** Briefly describe the kernel hook, eBPF helper/map, Go API, or user-space integration involved in the requested change, in beginner-friendly language.
3. **Make one small change at a time.** Prefer the smallest working vertical slice over broad refactoring or adding many features at once.
4. **Connect theory to practice.** When useful, identify the topic to read in _Learning eBPF_; don't invent quotations or page numbers.
5. **Show the complete data path.** For tracing features, explain where the event originates, what runs in kernel space, how data reaches Go, and how it is rendered. For integrations, identify where data is selected, transformed, and sent.
6. **Make verification reproducible.** Provide exact build, test, run, and manual event-generation commands, indicating whether each runs on the host or in the VM.
7. **Report limitations.** State what was actually executed and what could not be verified, including missing privileges, dependencies, kernel support, or external services.
8. **Preserve the learning experience.** Favor understandable code and focused explanations over complex abstractions. If asked to teach, use questions or checkpoints instead of immediately introducing the entire next milestone.
9. **Respect the milestone boundary.** Do not introduce the decision model, metrics exporter, Slack client, MCP server, or background agent architecture into the first tracer unless explicitly requested.

## Go standards

- Write idiomatic, readable Go; use `gofmt` and check `go vet` when available.
- Keep `main` thin. Encapsulate eBPF loading, attachment, event reading, and shutdown in internal packages.
- Return useful errors with context using `%w`; do not silently ignore important failures.
- Close links, maps, ring-buffer readers, and other resources. Handle `SIGINT`/`SIGTERM` cleanly.
- Use `context.Context` for controllable long-running operations where appropriate.
- Add tests for event decoding, formatting, filtering, and other pure Go logic without requiring root or a running eBPF program.
- Do not add frameworks, databases, or dependencies without a clear need.
- When adding asynchronous processing later, use bounded queues and well-defined behavior under backpressure. Do not block kernel-event consumption indefinitely on model inference, Slack delivery, or slow clients.
- Keep optional integrations disabled or inactive by default until explicitly configured. Failure of a future external integration should not silently prevent local observation.

## eBPF/C standards

- Keep programs short, event-driven, and compatible with the verifier.
- Prefer stable tracepoints for initial tracing; justify kprobes or other less stable hooks when needed.
- Use permitted eBPF helpers and bounded memory access. Never assume ordinary C library functions are available inside eBPF programs.
- Keep a documented, fixed event structure shared between C and Go; verify field sizes, alignment, and encoding when decoding.
- Distinguish process ID (TGID), thread ID (TID), and command name (`comm`); do not call `comm` a full executable path.
- Avoid expensive work per event. Filter in kernel space only when it reduces overhead without complicating the learning objective.
- When adding a ring buffer, handle reservation/output failures and explain that events can be lost under load. Expose drop counts when practical.
- Do not attach invasive or high-overhead probes globally without considering performance and operational safety.

## Future integration rules (apply only when requested)

### System Decision Model

- Keep inference and decision logic in user space, outside eBPF. The Go Agent may pass selected events or derived summaries to the model and receive results back.
- Define the event input and decision output before implementation. A future result might contain an event/reference ID, classification, anomaly indicator, severity, confidence (when meaningful), and a short rationale; these are **proposed fields**, not an existing contract.
- Treat output as an untrusted recommendation. Validate structure, handle timeouts and failures, and avoid interpreting model confidence as proof of malicious behavior.
- Prefer an explainable baseline (for example, transparent rules or simple statistical thresholds) before introducing opaque ML/LLM dependencies.
- Keep data local by default. Do not send kernel observations to hosted model APIs without explicit consent and an explicit privacy review.
- Never let model output directly execute commands, alter eBPF attachments, change policy, block processes, or trigger destructive actions.

### Prometheus

- Expose aggregate, low-cardinality metrics via a scrape endpoint only when this milestone is requested.
- Suitable future examples: observed events total, decode errors, ring-buffer delivery failures or known drops, and agent health. Do not invent drop counts that the implementation cannot measure.
- Avoid process IDs, executable paths, command lines, or other unbounded/sensitive values as metric labels.
- Make the endpoint's bind address and access exposure explicit; do not assume it should be publicly reachable.

### Slack alerts

- Alerts must be deliberate and configurable, based on explicit rules or vetted decision-model results—not every event.
- Add severity thresholds, deduplication/rate limits, and clear alert context when warranted; avoid alert storms.
- Keep credentials out of source control and logs. Redact secrets and sensitive process or file details from notifications.
- Handle delivery failure without disrupting the tracer. Do not claim an alert was delivered unless the API response confirms it.

### MCP support

- Treat MCP as a separate, optional interface to **existing** observations and summaries, not as an alternative eBPF event transport.
- Start with explicit, read-only operations, minimal permissions, bounded query results, and appropriate authentication/access controls for the selected transport.
- Do not expose arbitrary host commands, privileged eBPF operations, secret-bearing data, or uncontrolled model-generated actions.
- Document what external agents can access and how collected telemetry is filtered or redacted.

## Security and operational safety

- Treat privileged commands as VM-only by default. Never change the host kernel, host networking, system services, or security configuration without explicit approval.
- Do not disable SELinux, signature checks, or kernel security protections as a shortcut.
- Prefer read-only observation. Never make covert data collection, credential capture, or process tampering part of KernelScope.
- Keep sample output free of sensitive contents (arguments, paths, or payloads can expose secrets). Explain the privacy implications of collected events.
- Use a disposable VM for experiments that can destabilize the system. Separate demonstrations from production-readiness claims.
- Treat observed event fields as untrusted input. In later model, Slack, or MCP integrations, apply data minimization, redaction, bounded retention, and explicit user configuration before data leaves the process or machine.
- Never interpret future classification or anomaly findings as authorization for enforcement.

## Validation checklist for changes

- [ ] The change addresses only the requested milestone.
- [ ] New Go code is formatted and relevant unit tests pass (`go test ./...`).
- [ ] Any changed eBPF code compiles via the project's documented `bpf2go`/Clang workflow.
- [ ] Loading and attaching are tested inside the Rocky Linux VM when possible.
- [ ] A manual command reproduces the event (for example, executing `ls` or `curl`).
- [ ] All resources detach/close when KernelScope exits.
- [ ] Errors and dropped events are visible or documented.
- [ ] Documentation explains what was learned and any environment requirements.
- [ ] If a future integration is part of the requested change, its opt-in behavior, failure path, data exposure, and test method are documented.

## Definition of done: first milestone

KernelScope can attach a Go-managed eBPF program to `sched/sched_process_exec` in the Rocky Linux VM, observe newly executed commands, decode structured records in Go, print their timestamps and process identifiers, and detach cleanly on exit. A small experiment using `bpf_printk` may precede the structured-event version.

**Prometheus, Slack, MCP, and the System Decision Model are not acceptance criteria for this milestone.**

## Guiding rule

Learn the kernel concept, build the smallest working experiment, then integrate it into KernelScope. Grow toward the Go Agent and decision-support architecture incrementally, without sacrificing local observability, explainability, or safety. Keep Kubernetes out of the project until the maintainer explicitly requests it.
