# Little-Endian and Big-Endian

**Little-endian is a way computers arrange multi-byte data in memory.**

A computer stores data in bytes. When a value requires multiple bytes, the CPU architecture determines how those bytes are arranged in memory.

There are two common byte orders:

- **Little-endian:** Least significant byte first.
- **Big-endian:** Most significant byte first.

This concept is called **endianness**.

[![Big Endian vs Little Endian](https://assets.bytebytego.com/diagrams/0084-big-endian-vs-little-endian.png)](https://bytebytego.com/guides/big-endian-vs-little-endian/)

## 1. Simple Example

Consider a 32-bit hexadecimal number:

```text
0x12345678
```

This number occupies **4 bytes**.

| Byte   | Hexadecimal Value |
| ------ | ----------------- |
| Byte 1 | `12`              |
| Byte 2 | `34`              |
| Byte 3 | `56`              |
| Byte 4 | `78`              |

Each pair of hexadecimal digits represents one byte (8 bits).

### Little-Endian

In little-endian systems, the **least significant byte is stored at the lowest memory address**.

| Memory Address | Stored Byte |
| -------------- | ----------- |
| `1000`         | `78`        |
| `1001`         | `56`        |
| `1002`         | `34`        |
| `1003`         | `12`        |

Memory representation:

```text
Address:  1000   1001   1002   1003
          +------+------+------+------+
Value:    |  78  |  56  |  34  |  12  |
          +------+------+------+------+
            Low Address → High Address
```

### Big-Endian

In big-endian systems, the **most significant byte is stored at the lowest memory address**.

| Memory Address | Stored Byte |
| -------------- | ----------- |
| `1000`         | `12`        |
| `1001`         | `34`        |
| `1002`         | `56`        |
| `1003`         | `78`        |

Memory representation:

```text
Address:  1000   1001   1002   1003
          +------+------+------+------+
Value:    |  12  |  34  |  56  |  78  |
          +------+------+------+------+
            Low Address → High Address
```

**Important:** The numerical value `0x12345678` does not change. Only the order in which its bytes are stored changes.

## 2. Why Does This Matter for eBPF?

In our **KernelScope** project, `bpf2go` can generate two eBPF object files.

| File           | Meaning                               |
| -------------- | ------------------------------------- |
| `exec_bpfel.o` | eBPF object for little-endian systems |
| `exec_bpfeb.o` | eBPF object for big-endian systems    |

The suffixes mean:

- `bpfel` — BPF Little-Endian
- `bpfeb` — BPF Big-Endian

Our x86-64 Rocky Linux VM uses **little-endian**, so the little-endian eBPF object is selected.

### Why Two Object Files?

Different CPU architectures can use different byte orders.

An eBPF object must use the correct byte order for the target system.

Generating both variants allows the Go application to use the appropriate object for its target architecture.

## 3. Test Endianness Using Go

We can demonstrate byte ordering directly using Go.

```go
package main

import (
    "encoding/binary"
    "fmt"
)

func main() {
    value := uint32(0x12345678)

    little := make([]byte, 4)
    big := make([]byte, 4)

    binary.LittleEndian.PutUint32(little, value)
    binary.BigEndian.PutUint32(big, value)

    fmt.Printf("Little-endian: % x\n", little)
    fmt.Printf("Big-endian:    % x\n", big)
}
```

### Output

```text
Little-endian: 78 56 34 12
Big-endian:    12 34 56 78
```

### Explanation

The original value is:

```text
0x12345678
```

`binary.LittleEndian.PutUint32()` writes the least significant byte first.

`binary.BigEndian.PutUint32()` writes the most significant byte first.

Both functions represent the same numerical value using different byte orders.

## 4. Endianness in KernelScope Ring Buffers

Endianness matters when an eBPF program sends binary structures to a Go application through a BPF ring buffer.

For example, consider this eBPF structure:

```c
struct event {
    __u32 pid;
    __u32 uid;
};
```

Suppose:

```text
pid = 1000
uid = 1001
```

When the eBPF program writes this structure into the ring buffer, its integer fields are represented using the target system's byte order.

On our little-endian Rocky Linux VM:

```text
PID 1000 = 0x000003E8
```

The bytes are stored as:

```text
E8 03 00 00
```

When Go reads these bytes, it must interpret them using the correct byte order.

```go
var event struct {
    PID uint32
    UID uint32
}

err := binary.Read(
    reader,
    binary.LittleEndian,
    &event,
)
```

Using the wrong byte order would produce incorrect numerical values.

**Important:** The Go structure must also match the eBPF event's field layout and padding.

## 5. Endianness in Networking

Network protocols generally use **network byte order**, which is big-endian.

This is relevant when KernelScope is extended to monitor:

- IP addresses
- TCP connections
- UDP packets
- Network ports

For example, consider TCP port `8080`.

```text
Decimal:     8080
Hexadecimal: 0x1F90
```

Its two-byte representations are:

| Byte Order    | Bytes   |
| ------------- | ------- |
| Little-endian | `90 1F` |
| Big-endian    | `1F 90` |

Because TCP and UDP port fields use network byte order, their bytes appear in big-endian order in network packet headers.

In eBPF, helpers such as `bpf_ntohs()` and `bpf_htons()` can convert between network and host byte order.

---

**Little-endian means the least significant byte comes first in memory.**

It does not reverse the numerical value itself — only the way its bytes are represented.

For KernelScope, understanding endianness helps ensure that **binary data sent from eBPF programs is interpreted correctly by our Go application**.
