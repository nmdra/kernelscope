package main

import (
	"context"
	"fmt"

	"github.com/cilium/ebpf/link"
)

func watchExec(ctx context.Context) error {
    var objs execObjects

    if err := loadExecObjects(&objs, nil); err != nil {
        return fmt.Errorf("load exec tracer: %w", err)
    }
    defer objs.Close()

    tp, err := link.Tracepoint(
        "sched",
        "sched_process_exec",
        objs.TraceExec,
        nil,
    )
    if err != nil {
        return fmt.Errorf("attach exec tracer: %w", err)
    }
    defer tp.Close()

    fmt.Println("KernelScope: watching process executions")
    fmt.Println("Press Ctrl+C to stop")

    // Phase 1 still writes events with bpf_printk().
    <-ctx.Done()
    return nil
}
