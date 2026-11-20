package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
)

// watchFiles loads our eBPF program, attaches it,
// and consumes file-open events from the ring buffer.
func watchFiles(ctx context.Context) error {

	// STEP 1: Load generated eBPF objects.
	var objs fileObjects

	if err := loadFileObjects(&objs, nil); err != nil {
		return fmt.Errorf("load file tracer: %w", err)
	}
	defer objs.Close()

	// STEP 2: Attach to the openat syscall tracepoint.
	tp, err := link.Tracepoint(
		"syscalls",
		"sys_enter_openat",
		objs.TraceOpenat,
		nil,
	)
	if err != nil {
		return fmt.Errorf("attach tracepoint: %w", err)
	}
	defer tp.Close()

	// STEP 3: Open our kernel ring buffer for reading.
	reader, err := ringbuf.NewReader(objs.Events)
	if err != nil {
		return fmt.Errorf("open ring buffer: %w", err)
	}
	defer reader.Close()

	/*
	* reader.Read() blocks while waiting for events.
	*
	* When Ctrl+C cancels the context, closing the
	* reader unblocks the pending Read() call.
	*/
	stopReader := make(chan struct{})

	go func() {
		select {
			case <-ctx.Done():
			_ = reader.Close()
			case <-stopReader:
		}
	}()
	defer close(stopReader)

	fmt.Println("KernelScope: watching file-open attempts")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Printf(
		"%-8s %-7s %-16s %s\n",
		"TIME", "PID", "PROCESS", "FILE",
	)

	// STEP 4: Continuously receive kernel events.
	for {
		record, err := reader.Read()

		if err != nil {
			if errors.Is(err, ringbuf.ErrClosed) &&
			ctx.Err() != nil {
				return nil
			}

			return fmt.Errorf("read ring buffer: %w", err)
		}

		/*
		* Decode binary data into the Go structure
		* generated from our C struct file_event.
		*
		* Rocky Linux x86-64 uses little-endian.
		*/
		var event fileFileEvent

		err = binary.Read(
			bytes.NewReader(record.RawSample),
			binary.LittleEndian,
			&event,
		)
		if err != nil {
			log.Printf("decode event: %v", err)
			continue
		}

		// Convert null-terminated byte arrays to strings.
		comm := strings.TrimRight(
			string(event.Comm[:]), "\x00",
		)
		filename := strings.TrimRight(
			string(event.Filename[:]), "\x00",
		)

		if strings.HasPrefix(filename, "/proc/") {
			continue
		}

		// Display the received event.
		fmt.Printf(
			"%-8s %-7d %-16s %s\n",
			time.Now().Format("15:04:05"),
			event.Pid,
			comm,
			filename,
		)
	}
}
