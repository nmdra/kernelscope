# KernelScope

> [!NOTE]
> KernelScope is an experimental, ongoing eBPF learning project.

KernelScope is a Linux-first observability learning project built with Go and
eBPF. Its goal is to make process execution, file activity, network connections,
and performance events visible from the command line.

_Learning eBPF_ by Liz Rice is the theory reference.

## Current status

The repository currently provides a Rocky Linux 10.2 development VM through
Vagrant and KVM/libvirt. The Go CLI and eBPF tracer are not implemented yet.

The first milestone is a process-execution tracer attached to the
`sched/sched_process_exec` tracepoint. This kernel hook fires when a process
successfully executes a new program. A small eBPF program will capture a
structured event, and Go will read and display it.

The intended CLI is:

```text
sudo ./kernelscope watch exec
```

This command is a planned interface, not a working command yet.

## Requirements

- Linux host with KVM/libvirt configured
- Vagrant with the `vagrant-libvirt` plugin
- rsync

## Quick start

```bash
vagrant plugin install vagrant-libvirt
vagrant up --provider=libvirt
vagrant ssh
```

The VM uses 2 CPUs and 2 GiB of RAM. Project files sync to
`/home/vagrant/kernelscope` in the guest. Run `vagrant rsync` after local changes,
or `vagrant rsync-auto` to watch for changes.

Inside the VM, enter the synced project directory:

```bash
cd /home/vagrant/kernelscope
```

Run eBPF compilation, loading, and kernel tests inside the VM. The current
`Vagrantfile` does not provision Go, Clang/LLVM, or eBPF development dependencies.
Check installed versions and kernel capabilities before adding them. For
example, inside the VM:

```bash
uname -r
ls -l /sys/kernel/btf/vmlinux
```

Kernel BTF describes kernel types and can support portable eBPF programs through
CO-RE (Compile Once, Run Everywhere).

On the host, stop the VM with `vagrant halt`.

## Planned architecture

```text
Linux tracepoint
    -> small C eBPF program
    -> ring buffer with structured events
    -> Go reader using github.com/cilium/ebpf
    -> CLI output
```

The Go process will load, attach, read from, and detach the eBPF program.
Kernel-side C will only collect event data when the hook fires. `bpf2go` will
compile the C program and generate its Go bindings.

Early experiments may use `bpf_printk` and `trace_pipe` to understand the data
path. Integrated features will use structured events rather than tracing logs.

## Milestones

1. **Process execution:** timestamped process events from
   `sched/sched_process_exec`, structured decoding in Go, and clean shutdown.
2. **File activity:** observe file operations.
3. **TCP connections:** show connection activity.
4. **Syscall latency and performance:** measure selected kernel events.
5. **Interactive terminal UI:** explore observations in the terminal.
6. **Incident timeline:** organize related observations over time.

Kubernetes, orchestration integrations, databases, cloud backends, and complex
frontends are outside the initial scope. KernelScope starts with read-only
observation, not blocking or enforcement.

## Development approach

- Build one small, testable feature at a time and explain its kernel-to-Go data
  path.
- Add tests for pure Go logic, such as event decoding and formatting, without
  requiring root or an attached eBPF program.
- Document exact build and verification commands as each feature is added.
- Use the disposable VM for privileged experiments. Do not disable SELinux or
  kernel security protections to make an experiment work.
- Treat collected command names, paths, arguments, and payloads as potentially
  sensitive. Keep sample output free of secrets.

See [AGENTS.md](AGENTS.md) for agent instructions, coding conventions, safety
rules, and the first milestone's definition of done.

## Troubleshooting

If Vagrant waits for an IP address on a host using UFW, see the
[libvirt DHCP troubleshooting guide](docs/vagrant/trobleshoot/vagrant-libvirt-dhcp-ufw-rca.md).

Local VM state and SSH keys in `.vagrant/` are not committed.
