# CHIP-8 Systems Emulator

> A cycle-accurate CHIP-8 emulator written in Go — featuring a real-time debug panel, disassembler, metrics reporter, and a clean concurrency model.

```
┌──────────────────────────────────────────┐
│ CHIP-8 Systems Emulator v1.0             │
│ ROM:   roms/tetris.ch8                   │
│ Speed: 500  Hz  Debug: false             │
│ Keys:  1234/QWER/ASDF/ZXCV  ESC=pause   │
└──────────────────────────────────────────┘
```

---

## Table of Contents

1. [What is CHIP-8?](#1-what-is-chip-8)
2. [What is an Emulator?](#2-what-is-an-emulator)
3. [Project Features](#3-project-features)
4. [Architecture Overview](#4-architecture-overview)
5. [How CHIP-8 Hardware Works](#5-how-chip-8-hardware-works)
6. [Codebase Walkthrough](#6-codebase-walkthrough)
   - [cpu/chip8.go — The CPU Core](#cpuchip8go--the-cpu-core)
   - [cpu/disasm.go — The Disassembler](#cpudisasmgo--the-disassembler)
   - [cpu/metrics.go — The Metrics Reporter](#cpumetricsgo--the-metrics-reporter)
   - [gui/gui.go — The Renderer & Debug Panel](#guiguigo--the-renderer--debug-panel)
   - [cli/cli.go — The CLI Parser](#cliclogo--the-cli-parser)
   - [cmd/chip8/main.go — The Entrypoint](#cmdchip8maingo--the-entrypoint)
7. [The Fetch-Decode-Execute Cycle](#7-the-fetch-decode-execute-cycle)
8. [Concurrency Model](#8-concurrency-model)
9. [Opcode Reference](#9-opcode-reference)
10. [CHIP-8 Quirks Implemented](#10-chip-8-quirks-implemented)
11. [Getting Started](#11-getting-started)
12. [Controls](#12-controls)
13. [Project Structure](#13-project-structure)
14. [Design Decisions & Trade-offs](#14-design-decisions--trade-offs)

---

## 1. What is CHIP-8?

CHIP-8 is an **interpreted programming language** created in the mid-1970s by Joseph Weisbecker. It was designed to make game development easier on early microcomputers like the COSMAC VIP and Telmac 1800.

Think of CHIP-8 not as a physical chip, but as a **virtual machine specification** — a minimal, well-defined set of rules describing:

- How much memory a program can use (4KB)
- What registers are available (16 general-purpose)
- What instructions the CPU understands (35 opcodes)
- What the display looks like (64×32 pixels, monochrome)
- How sound and input work

Because the spec is tiny, CHIP-8 has become the **"Hello, World" of emulator development**. Writing a CHIP-8 emulator teaches you every fundamental concept — memory, registers, the fetch-decode-execute cycle, display rendering, and input handling — before you tackle something complex like a Game Boy or NES.

---

## 2. What is an Emulator?

An emulator is a **software program that replicates the behavior of one computing system on another**. It doesn't run original hardware — it simulates it in software, instruction by instruction.

The core loop of any emulator looks like this:

```
while running:
    opcode = fetch(memory[PC])   // Read next instruction
    PC += 2                       // Advance program counter
    decode(opcode)                // Figure out what it means
    execute(opcode)               // Do the thing
```

This project implements exactly that loop for the CHIP-8 virtual machine, plus rendering, input, sound timers, and a live debug panel.

---

## 3. Project Features

| Feature | Details |
|---|---|
| **Full Opcode Coverage** | All 35 standard CHIP-8 opcodes implemented |
| **Accurate Timers** | Delay and sound timers decrement at exactly 60 Hz |
| **Configurable CPU Speed** | Default 500 Hz, adjustable via CLI flag |
| **Debug Panel** | Live register viewer, memory hex dump, opcode disassembler |
| **CLI Disassembler** | Dump any ROM to annotated human-readable assembly |
| **Metrics Reporter** | Real-time IPS (instructions per second) + jitter tracking |
| **Pause/Resume** | ESC key pauses the entire emulator |
| **CHIP-8 Quirks** | VF-reset quirk for bitwise ops, clipping (no wrapping) |
| **Memory Viewer** | Scrollable hex dump with PC highlighted live |
| **44 Test ROMs** | Comprehensive suite covering every opcode |

---

## 4. Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│                    main.go                          │
│  Parses CLI → Spawns goroutines → Starts Ebiten     │
└──────┬──────────────────────────────────────────────┘
       │
       ├──── Goroutine: CPU tick (500 Hz)
       │         └── c.Step() → c.Execute(opcode)
       │
       ├──── Goroutine: Timer tick (60 Hz)
       │         └── DelayTimer--, SoundTimer--
       │
       ├──── Goroutine: Metrics reporter (1 Hz)
       │         └── Reads InstrCount, logs IPS
       │
       └──── Ebiten main loop (60 Hz, main thread)
                 ├── Update() → poll keys
                 └── Draw()   → render display + debug panel
```

Every goroutine that touches CPU state acquires `c.Mu` (a `sync.Mutex`) first. This prevents data races between the CPU goroutine and the GUI render goroutine.

---

## 5. How CHIP-8 Hardware Works

Understanding the virtual hardware is the key to understanding every line of code in this emulator.

### Memory — 4096 bytes

CHIP-8 has a flat 4KB address space (`0x000`–`0xFFF`):

```
0x000 ┌──────────────────┐
      │  Interpreter     │  (reserved, not used by ROMs)
0x050 ├──────────────────┤
      │  Font sprites    │  Built-in 0–F digit bitmaps (80 bytes)
0x200 ├──────────────────┤
      │                  │
      │  ROM / Program   │  All ROMs load here
      │                  │
0xFFF └──────────────────┘
```

ROMs are loaded at address `0x200`. The program counter (PC) starts there.

### Registers

| Register | Size | Purpose |
|---|---|---|
| `V0`–`VF` | 8-bit | General-purpose. VF is also used as a flag register (carry, borrow, collision) |
| `I` | 16-bit | Index register — points to a memory address (used for sprites, BCD, etc.) |
| `PC` | 16-bit | Program Counter — address of the next instruction to execute |
| `SP` | 8-bit | Stack Pointer — index into the 16-level call stack |
| `DelayTimer` | 8-bit | Counts down at 60 Hz. Readable by programs |
| `SoundTimer` | 8-bit | Counts down at 60 Hz. Beeps while non-zero |

### Display

64×32 pixels, monochrome (on/off). Sprites are drawn with **XOR logic**: drawing over an existing pixel turns it off, and sets `VF = 1` (collision flag). This is how games detect when a bullet hits a sprite.

### Stack

A 16-level deep call stack. `CALL` pushes the return address; `RET` pops it. This is how CHIP-8 implements subroutines.

### Input

16 keys (0–F), polled every frame. No interrupts — programs busy-poll or use the blocking `FX0A` wait-for-key instruction.

### Font Sprites

Each digit (0–F) has a pre-built 5-byte sprite stored at `0x050`–`0x09F`. Programs use `FX29` to point `I` at the sprite for digit `Vx`, then draw it.

---

## 6. Codebase Walkthrough

### `cpu/chip8.go` — The CPU Core

This is the heart of the emulator. It defines the `Chip8` struct and the `Execute` method.

**The struct:**

```go
type Chip8 struct {
    Memory     [4096]uint8      // Flat 4KB address space
    V          [16]uint8        // General-purpose registers V0–VF
    I          uint16           // Index register
    PC         uint16           // Program counter (starts at 0x200)
    Stack      [16]uint16       // Subroutine call stack
    SP         uint16           // Stack pointer
    DelayTimer uint8
    SoundTimer uint8
    Display    [64 * 32]uint8   // Flat pixel buffer (1 = on, 0 = off)
    Keys       [16]bool         // Key state (true = pressed)
    Mu         sync.Mutex       // Guards all state for concurrent access
    InstrCount atomic.Int64     // Lock-free instruction counter for metrics
    LastOpcode uint16           // Last opcode, for debug panel display
    DrawFlag   bool             // True when display needs re-rendering
}
```

**`Step()` — one CPU cycle:**

```go
func (c *Chip8) Step() {
    if c.PC >= 4094 { return }           // Bounds check
    opcode := uint16(c.Memory[c.PC])<<8  // Fetch high byte
             | uint16(c.Memory[c.PC+1])  // Fetch low byte → combine into 16-bit opcode
    c.PC += 2                            // Advance past this instruction
    c.LastOpcode = opcode
    c.InstrCount.Add(1)                  // Atomic: no lock needed
    c.Execute(opcode)
}
```

**Why `<<8`?** CHIP-8 opcodes are 2 bytes, stored big-endian. If memory holds `0xD0` and `0x15`, the opcode is `0xD015`. Shifting the high byte left by 8 bits and OR-ing in the low byte reconstructs it.

**`Execute()` — the decoder:**

The switch statement on `opcode & 0xF000` extracts the top nibble (4 bits), which identifies the instruction family. Sub-cases handle the rest:

```go
x   := (opcode & 0x0F00) >> 8   // Second nibble: usually a register index (0–F)
y   := (opcode & 0x00F0) >> 4   // Third nibble: usually a register index
n   := opcode & 0x000F          // Fourth nibble: small constant
nn  := uint8(opcode & 0x00FF)   // Lower byte: 8-bit immediate value
nnn := opcode & 0x0FFF          // Lower 12 bits: memory address
```

This bit-masking pattern is the bread and butter of emulator development.

---

### `cpu/disasm.go` — The Disassembler

A disassembler does the reverse of a CPU: instead of executing an opcode, it converts it to a **human-readable string**.

```go
func Disassemble(addr uint16, opcode uint16) string {
    // Same nibble extraction as Execute()
    switch opcode & 0xF000 {
    case 0xD000:
        return fmt.Sprintf("DRW   V%X, V%X, %d", x, y, n)
    // ...
    }
}
```

`DisasmROM` does a two-pass scan:
- **Pass 1:** Collect all jump/call targets → generate labels (`L0208:`, etc.)
- **Pass 2:** Disassemble each instruction and annotate (e.g., `; <-- infinite loop`)

This is the same technique used by tools like `objdump` and IDA Pro.

---

### `cpu/metrics.go` — The Metrics Reporter

Runs in its own goroutine, sampling `InstrCount` every second. Because `InstrCount` is an `atomic.Int64`, no mutex is needed to read it — atomic operations are safe across goroutines by definition.

```go
current := c.InstrCount.Load()
ips := current - lastCount   // Instructions executed in last second
```

It also tracks **jitter** — the deviation from the expected 1000ms tick interval. High jitter means the OS scheduler is causing timing instability, which would cause the emulator to run at inconsistent speed.

---

### `gui/gui.go` — The Renderer & Debug Panel

Built on [Ebitengine](https://ebitengine.org/), a 2D game engine for Go. Ebiten calls `Update()` and `Draw()` at 60 Hz on the main thread.

**Rendering the display:**

```go
// The CHIP-8 display is a flat [64*32]uint8 array.
// We convert it to an RGBA image and upscale 15× for the window.
img := image.NewRGBA(image.Rect(0, 0, ChipW, ChipH))
for i, pixel := range displayCopy {
    x := i % 64
    y := i / 64
    if pixel == 1 {
        img.SetRGBA(x, y, colPixelOn)   // Matrix green
    } else {
        img.SetRGBA(x, y, colPixelOff)  // Very dark tint
    }
}
g.offscreen.WritePixels(img.Pix)
// Draw scaled 15× onto the screen
op.GeoM.Scale(15, 15)
screen.DrawImage(g.offscreen, op)
```

**Why copy the display before rendering?**

The CPU goroutine can modify `Display` at any time. Taking a snapshot under the mutex, then rendering the copy outside the mutex, means the CPU is never blocked waiting for a slow draw call to finish.

**The debug panel** renders:
- PC, SP, I, timers (from `Snapshot`)
- V0–VF register values
- Last opcode disassembled live
- Scrollable hex memory dump with PC highlighted
- Active key state

---

### `cli/cli.go` — The CLI Parser

Parses `os.Args` into a `Config` struct. Supports three commands:

```
chip8 run <rom> [--speed N] [--debug]
chip8 disasm <rom>
chip8 version | help
```

No external CLI library — pure standard library argument parsing.

---

### `cmd/chip8/main.go` — The Entrypoint

Wires everything together. Key responsibility: spawn goroutines and hand off to Ebiten.

The `done` channel is closed when Ebiten exits, which causes all goroutines to return from their `select` loop cleanly — a standard Go shutdown pattern.

---

## 7. The Fetch-Decode-Execute Cycle

This is the fundamental loop of every CPU — real or emulated.

```
  ┌──────────────────────────────────────────────┐
  │                                              │
  ▼                                              │
FETCH:   Read 2 bytes from memory[PC]; PC += 2   │
  │                                              │
  ▼                                              │
DECODE:  Extract nibbles x, y, n, nn, nnn        │
  │                                              │
  ▼                                              │
EXECUTE: Perform the operation (write register,  │
         jump, draw sprite, etc.)                │
  │                                              │
  └──────────────────────────────────────────────┘
```

In this emulator, `Step()` handles Fetch and bookkeeping; `Execute()` handles Decode and Execute.

A real CPU does this billions of times per second in hardware. This emulator does it 500 times per second in software — slow enough that a `time.Ticker` goroutine drives it at the correct rate.

---

## 8. Concurrency Model

Go's goroutines make it easy to run the CPU, timers, metrics, and GUI simultaneously — but they all share the same `Chip8` struct, which creates data race risk.

```
                    ┌─────────────┐
Goroutine: CPU ───► │             │ ◄─── Goroutine: Timers
                    │  Chip8 Mu   │
Goroutine: GUI ───► │  (Mutex)    │ ◄─── (metrics uses atomic, no lock)
                    └─────────────┘
```

**Rules:**
- Any goroutine reading or writing `Memory`, `V`, `PC`, `Display`, `Keys`, `DelayTimer`, `SoundTimer` must hold `c.Mu`.
- `InstrCount` uses `sync/atomic` — safe to read/write without a mutex, from any goroutine.
- `GetIPS()` reads `lastIPS` which is written only by the metrics goroutine. This is technically a benign data race (integers are word-sized), but could be made atomic for strict correctness.

The GUI takes a **snapshot** under the lock, then renders the snapshot without holding the lock — keeping the critical section as short as possible so the CPU goroutine is never blocked for long.

---

## 9. Opcode Reference

Every CHIP-8 instruction and its implementation:

| Opcode | Mnemonic | What it does |
|---|---|---|
| `00E0` | `CLS` | Clear the display |
| `00EE` | `RET` | Return from subroutine (pop stack) |
| `1NNN` | `JP NNN` | Jump to address NNN |
| `2NNN` | `CALL NNN` | Call subroutine at NNN (push PC to stack) |
| `3XNN` | `SE Vx, NN` | Skip next instruction if Vx == NN |
| `4XNN` | `SNE Vx, NN` | Skip if Vx != NN |
| `5XY0` | `SE Vx, Vy` | Skip if Vx == Vy |
| `6XNN` | `LD Vx, NN` | Set Vx = NN |
| `7XNN` | `ADD Vx, NN` | Set Vx = Vx + NN (no carry flag) |
| `8XY0` | `LD Vx, Vy` | Set Vx = Vy |
| `8XY1` | `OR Vx, Vy` | Vx \|= Vy; VF = 0 |
| `8XY2` | `AND Vx, Vy` | Vx &= Vy; VF = 0 |
| `8XY3` | `XOR Vx, Vy` | Vx ^= Vy; VF = 0 |
| `8XY4` | `ADD Vx, Vy` | Vx += Vy; VF = carry |
| `8XY5` | `SUB Vx, Vy` | Vx -= Vy; VF = NOT borrow |
| `8XY6` | `SHR Vx` | Vx >>= 1; VF = shifted-out bit |
| `8XY7` | `SUBN Vx, Vy` | Vx = Vy - Vx; VF = NOT borrow |
| `8XYE` | `SHL Vx` | Vx <<= 1; VF = shifted-out bit |
| `9XY0` | `SNE Vx, Vy` | Skip if Vx != Vy |
| `ANNN` | `LD I, NNN` | Set I = NNN |
| `BNNN` | `JP V0, NNN` | Jump to NNN + V0 |
| `CXNN` | `RND Vx, NN` | Vx = random byte & NN |
| `DXYN` | `DRW Vx, Vy, N` | Draw N-row sprite at (Vx,Vy); VF = collision |
| `EX9E` | `SKP Vx` | Skip if key Vx is pressed |
| `EXA1` | `SKNP Vx` | Skip if key Vx is NOT pressed |
| `FX07` | `LD Vx, DT` | Vx = delay timer |
| `FX0A` | `LD Vx, K` | Wait for keypress; Vx = key (blocking) |
| `FX15` | `LD DT, Vx` | Delay timer = Vx |
| `FX18` | `LD ST, Vx` | Sound timer = Vx |
| `FX1E` | `ADD I, Vx` | I += Vx |
| `FX29` | `LD F, Vx` | I = address of font sprite for digit Vx |
| `FX33` | `LD B, Vx` | Store BCD of Vx at I, I+1, I+2 |
| `FX55` | `LD [I], Vx` | Store V0–Vx in memory starting at I |
| `FX65` | `LD Vx, [I]` | Load V0–Vx from memory starting at I |

**What is BCD?** Binary-Coded Decimal. `FX33` breaks Vx into its decimal digits. For example, `Vx = 123` stores `1` at `Memory[I]`, `2` at `Memory[I+1]`, `3` at `Memory[I+2]`. Games use this to display numeric scores.

**The blocking wait (`FX0A`):** Implemented by re-winding the PC by 2 if no key is pressed, causing `Step()` to re-execute the same instruction next cycle. This is the standard CHIP-8 technique for blocking operations.

---

## 10. CHIP-8 Quirks Implemented

"Quirks" are behaviors where different CHIP-8 interpreters disagreed. Getting them right is what separates a compatible emulator from a broken one.

| Quirk | Behavior | Why it matters |
|---|---|---|
| **VF reset** | `8XY1`, `8XY2`, `8XY3` set `VF = 0` after the operation | Original COSMAC VIP behavior. Many ROMs rely on this. |
| **Display clipping** | Sprites that go past the right or bottom edge are clipped, not wrapped | Some interpreters wrap; clipping matches the majority of ROMs. |
| **Carry/borrow flag timing** | VF is set *after* Vx is modified in `8XY4/5/7` | Correct order matters if Vx == VF. |

---

## 11. Getting Started

### Prerequisites

- Go 1.21+
- A system with OpenGL support (Linux/macOS/Windows)

### Build & Run

```bash
git clone https://github.com/yourname/chip8
cd chip8
go mod tidy

# Run a ROM
go run ./cmd/chip8 run roms/tetris.ch8

# Run at custom speed with debug panel
go run ./cmd/chip8 run roms/ibm.ch8 --speed 700 --debug

# Disassemble a ROM to stdout
go run ./cmd/chip8 disasm roms/ibm.ch8

# Build a binary
go build -o chip8 ./cmd/chip8
./chip8 run roms/invaders.ch8
```

### CLI Flags

```
chip8 run <rom_path> [options]

Options:
  --speed N     CPU speed in Hz (default: 500)
  --debug       Enable debug panel (registers, memory, disassembly)

chip8 disasm <rom_path>
chip8 version
chip8 help
```

---

## 12. Controls

CHIP-8 uses a 16-key hex keypad (0–F). This emulator maps it to a QWERTY keyboard:

```
CHIP-8 Keypad      Keyboard Mapping
┌───┬───┬───┬───┐  ┌───┬───┬───┬───┐
│ 1 │ 2 │ 3 │ C │  │ 1 │ 2 │ 3 │ 4 │
├───┼───┼───┼───┤  ├───┼───┼───┼───┤
│ 4 │ 5 │ 6 │ D │  │ Q │ W │ E │ R │
├───┼───┼───┼───┤  ├───┼───┼───┼───┤
│ 7 │ 8 │ 9 │ E │  │ A │ S │ D │ F │
├───┼───┼───┼───┤  ├───┼───┼───┼───┤
│ A │ 0 │ B │ F │  │ Z │ X │ C │ V │
└───┴───┴───┴───┘  └───┴───┴───┴───┘
```

| Key | Action |
|---|---|
| `ESC` | Pause / Resume |
| `Scroll Wheel` | Scroll memory viewer (debug mode) |
| `Page Up/Down` | Fast-scroll memory viewer (debug mode) |

---

## 13. Project Structure

```
chip8/
├── cmd/chip8/
│   └── main.go          # Entrypoint: wires CPU, GUI, goroutines
├── internal/
│   ├── cli/
│   │   └── cli.go       # CLI argument parsing, version string
│   ├── cpu/
│   │   ├── chip8.go     # CPU struct, Step(), Execute(), memory
│   │   ├── disasm.go    # Disassembler (single opcode + full ROM)
│   │   └── metrics.go   # IPS counter + jitter reporter goroutine
│   └── gui/
│       └── gui.go       # Ebiten renderer, debug panel, key input
└── roms/
    ├── tetris.ch8
    ├── invaders.ch8
    ├── ibm.ch8
    └── test_*.ch8       # 40+ test ROMs for individual opcodes
```

The `internal/` package boundary prevents external packages from importing CPU or GUI internals — standard Go project hygiene.

---

## 14. Design Decisions & Trade-offs

### Why separate goroutines for CPU and timers?

CHIP-8 specifies that the CPU runs at ~500 Hz and timers at exactly 60 Hz. These are independent rates. Using separate goroutines with `time.Ticker` gives each its own accurate clock without coupling them.

The alternative — running timers every N CPU steps — is simpler but less accurate, because timer decrement becomes a function of CPU speed rather than wall-clock time.

### Why Ebiten?

Ebiten handles the platform-specific OpenGL context, game loop timing, and input polling. It runs `Update()` and `Draw()` at a consistent 60 Hz, which maps perfectly to CHIP-8's display and timer rate. The alternative would be raw SDL2 bindings (more control, more boilerplate).

### Why is the display a `[64*32]uint8` and not `[64*32]bool`?

Convenience: `uint8` lets us write pixels directly as `0` or `1` and do XOR arithmetic naturally. A bool would require conversion at the draw site.

### Why `atomic.Int64` for `InstrCount` but `sync.Mutex` for everything else?

`InstrCount` is only ever incremented (written) by the CPU goroutine and read by the metrics goroutine. An atomic integer is the lowest-overhead synchronization primitive for this single-writer pattern. The rest of the CPU state has complex interdependencies (PC + Memory + V registers all change together in one instruction), so a mutex protecting the whole struct is simpler and safer than per-field atomics.

### Why clone the display before rendering?

Rendering can take non-trivial time, especially when building the debug panel. Holding the mutex during the entire render would starve the CPU goroutine of the lock. Taking a snapshot (cheap) under the lock and rendering the copy (slow) outside the lock keeps the critical section tiny.


