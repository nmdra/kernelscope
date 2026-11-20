// Kernel data types generated from BTF.
#include "vmlinux.h"

// Provides eBPF helpers such as bpf_get_current_comm()
// and macros such as SEC() and bpf_printk().
#include <bpf/bpf_helpers.h>

// Declares the eBPF program license.
// GPL-compatible licensing permits GPL-only BPF helpers.
char LICENSE[] SEC("license") = "GPL";

/*
 * SEC() identifies the kernel hook this program targets.
 *
 * sched_process_exec is triggered when a process
 * successfully executes a new program.
 *
 * The Go loader will attach this program to that hook.
 */
SEC("tracepoint/sched/sched_process_exec")
int trace_exec(void *ctx)
{
    /*
     * Allocate space for the current task's command name.
     *
     * TASK_COMM_LEN is 16 bytes, including the null
     * terminator, so long names may be truncated.
     */
    char comm[16];

    /*
     * Ask the kernel to copy the current task's
     * command name into our buffer.
     *
     * This is an eBPF helper function, not a normal
     * userspace function call.
     */
    bpf_get_current_comm(comm, sizeof(comm));

    /*
     * Print a debugging message to the kernel's
     * tracing buffer.
     *
     * Read these messages through trace_pipe.
     *
     * For production event streaming, we'll replace
     * bpf_printk() with an eBPF ring buffer.
     */
    bpf_printk("KernelScope: process executed: %s", comm);

    /*
     * Return successfully.
     *
     * For this tracing program, returning 0 simply
     * finishes this invocation. It does not control
     * whether the process is allowed to execute.
     */
    return 0;
}
