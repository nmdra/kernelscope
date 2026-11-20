package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
    // Cancel the context on Ctrl+C or SIGTERM.
    ctx, stop := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,
    )
    defer stop()

    if err := run(ctx, os.Args[1:]); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
}

// Route CLI commands to their corresponding tracers.
func run(ctx context.Context, args []string) error {
    if len(args) != 2 || args[0] != "watch" {
        return fmt.Errorf("usage: kernelscope watch [exec|files]")
    }

    switch args[1] {
    case "exec":
        return watchExec(ctx)
    case "files":
        return watchFiles(ctx)
    default:
        return fmt.Errorf("unknown tracer: %s", args[1])
    }
}
