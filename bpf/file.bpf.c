//go:build ignore

#include "vmlinux.h"
#include <bpf/bpf_helpers.h>

char LICENSE[] SEC("license") = "GPL";

/*
    * This structure represents one file-open event.
    *
    * It is the binary format shared between the
    * kernel-side eBPF program and our Go application.
    */
struct file_event {
    __u32 pid;          //  Process ID
    __u32 tid;          //  Thread ID
    __u32 flags;        //  openat() flags
    __u8 comm[16];      //  Process name
    __u8 filename[256]; //  File path
};

/*
    * Define and eBPF ring buffer map.
    *
    * The kernerl writes events here and the Go application
    * reads them from user space.
    *
    * Capacity: 4 MiB
    */
struct {
    __uint(type, BPF_MAP_TYPE_RINGBUF);
    __uint(max_entries, 1 << 22);
    __type(value, struct file_event);
} events SEC(".maps");

/*
    * Attach to the openat() syscall entry tracepoint.
    *
    * This function runs when a process attempts
    * to open a file through openat().
    */
SEC("tracepoint/syscalls/sys_enter_openat")
int trace_openat(struct trace_event_raw_sys_enter *ctx)
{
    // Zero-initialize an event on the eBPF stack.
    struct file_event event = {};

    // Upper 32 bits contain the process ID (TGID).
    __u64 pid_tgid = bpf_get_current_pid_tgid();
    event.pid = pid_tgid >> 32;
    event.tid = (__u32)pid_tgid;

    // Third openat() argument contains the open flags.
    event.flags = (__u32)ctx->args[2];

    // Get the current task's short command name.
    bpf_get_current_comm(event.comm, sizeof(event.comm));

    /*
        * openat(dirfd, filename, flags, mode)
        *
        * ctx->args[1] contains a pointer to the filename
        * in the calling process's user-space memory.
        *
        * The helper safely copies the string into
        * our event structure.
        */
    const char *filename = (const char *)ctx->args[1];

    long ret = bpf_probe_read_user_str(
        event.filename,
        sizeof(event.filename),
        filename
    );

    // If reading the filename fails, skip the event.
    if (ret < 0)
        return 0;

    /*
        * Copy this structured event into the ring buffer.
        *
        * Our Go application will consume it.
        * If the buffer is full, this event may be dropped.
        */
    bpf_ringbuf_output(&events, &event, sizeof(event), 0);

    return 0;
}
