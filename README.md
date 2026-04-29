# CHIP-8 Emulator

## Project Overview

CHIP-8 is a simple, interpreted programming language designed in the 1970s for creating video games on early microcomputers. It features a virtual machine with a small instruction set, making it an excellent starting point for learning about emulation and low-level systems.

This project is a CHIP-8 emulator written in Go. It aims to faithfully replicate the behavior of a CHIP-8 virtual machine, allowing users to load and run CHIP-8 ROMs. The emulator is currently **75–80% complete**, with most core features implemented.

### Current State
- **Supported Features**: CPU, memory, opcode handling, display, input, timers, and basic ROM execution.
- **Incomplete Features**: Some advanced opcodes, edge-case handling, and optimizations.
- **Planned Features**: Debugging tools, sound support, and performance improvements.

---

## Features

### Completed
- **CPU**: Emulation of the CHIP-8 processor, including registers, stack, and program counter.
- **Memory**: 4KB memory model with proper allocation for ROMs and reserved areas.
- **Opcode Handling**: Decoding and execution of most CHIP-8 instructions.
- **Display**: 64x32 monochrome pixel rendering.
- **Input**: Support for 16-key hexadecimal keypad.
- **Timers**: Implementation of delay and sound timers.

### In Progress
- **Sound**: Beeper functionality for sound instructions.
- **Edge Cases**: Handling rare or undocumented opcode behaviors.

### Planned
- **Debugging Tools**: Step-through execution, opcode inspection, and state dumping.
- **Performance Optimizations**: Faster rendering and input handling.

---

## Architecture & Design

### Codebase Structure
- **`cmd/chip8/main.go`**: Entry point for the emulator.
- **`internal/cpu/chip8.go`**: Core CPU implementation, including the fetch-decode-execute cycle.
- **`internal/cli/cli.go`**: Command-line interface for argument parsing and usage printing.
- **`roms/`**: Directory for CHIP-8 ROM files.

### Key Components
1. **CPU/Emulation Loop**:
   - The CPU fetches, decodes, and executes instructions in a continuous loop.
   - See `internal/cpu/chip8.go` for the `Run()` method.
2. **Opcode Decoding**:
   - Instructions are parsed using bitwise operations.
   - Example: `0x8XY4` (ADD) is decoded by masking and shifting bits.
3. **Memory Model**:
   - 4KB memory, with the first 512 bytes reserved for the interpreter.
   - ROMs are loaded starting at address `0x200`.
4. **Display Handling**:
   - 64x32 monochrome display, updated using XOR operations for drawing.
5. **Input Handling**:
   - 16-key keypad mapped to hexadecimal values.
6. **Timers**:
   - Delay and sound timers decrement at 60Hz.

---

## How It Works (Educational Section)

### Fetch → Decode → Execute Cycle
1. **Fetch**:
   - The CPU reads the next 2 bytes from memory at the program counter (`PC`).
2. **Decode**:
   - The instruction is decoded using bitwise operations.
   - Example: `0x8XY4` is identified as an ADD operation.
3. **Execute**:
   - The decoded instruction is executed, modifying the CPU state.

### Key Concepts
- **Opcode Parsing**:
  - Instructions are 2 bytes long and require masking/shifting to decode.
  - Example:
    ```go
    opcode := memory[pc] << 8 | memory[pc+1]
    ```
- **Stack and Subroutines**:
  - The stack is used for subroutine calls and returns.
  - Example: `CALL` pushes the current `PC` onto the stack.
- **Timers**:
  - Decremented at 60Hz, synchronized with the system clock.

### Diagram: Fetch → Decode → Execute
```
+---------+       +---------+       +---------+
|  Fetch  | --->  |  Decode | --->  | Execute |
+---------+       +---------+       +---------+
```

---

## Getting Started

### Prerequisites
- Install [Go](https://golang.org/) (version 1.20 or higher).

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/chip8.git
   cd chip8
   ```
2. Build the emulator:
   ```bash
   go build -o chip8 ./cmd/chip8
   ```

### Running the Emulator
1. Load a ROM:
   ```bash
   ./chip8 run roms/test.ch8
   ```
2. Disassemble a ROM:
   ```bash
   ./chip8 --disasm roms/test.ch8
   ```

---

## Code Walkthrough

### Key Functions
- **`cpu.New()`**:
  - Initializes the CPU, setting up memory, registers, and timers.
- **`cpu.LoadROM()`**:
  - Loads a ROM into memory starting at address `0x200`.
- **`cpu.Run()`**:
  - Implements the fetch-decode-execute cycle.

### Design Decisions
- **Bitwise Operations**:
  - Used for efficient opcode decoding.
- **Memory Layout**:
  - Mirrors the original CHIP-8 specification for compatibility.

---

## Project Structure

```
chip8/
├── cmd/
│   └── chip8/
│       └── main.go       # Entry point
├── internal/
│   ├── cli/
│   │   └── cli.go        # Command-line interface
│   └── cpu/
│       └── chip8.go      # CPU implementation
├── roms/                 # Test ROMs
│   └── *.ch8
└── go.mod                # Go module file
```

---

## Limitations & TODOs

### Limitations
- No sound support yet.
- Limited debugging tools.
- Some edge cases for opcodes are untested.

### TODOs
- Implement sound functionality.
- Add debugging tools.
- Optimize rendering and input handling.

---

## Contributing

We welcome contributions! To get started:
1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a clear description.

---

Thank you for exploring this CHIP-8 emulator! Whether you're here to learn or contribute, we hope you enjoy the journey into emulation.