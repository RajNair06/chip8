package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"chip8/internal/cpu"
)

const version = "1.0.0"

func printUsage() {
	executable := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "CHIP-8 Virtual Machine - Version %s\n\n", version)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <command> <rom_path>\n\n", executable)
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  run        Initialize and execute the specified ROM\n")
	fmt.Fprintf(os.Stderr, "  --disasm   Disassemble the specified ROM into human-readable opcodes\n")
	fmt.Fprintf(os.Stderr, "  --version  Show version information\n")
}

func main() {
	// 1. Basic Argument Validation
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle version flag early
	if command == "--version" || command == "-v" {
		fmt.Printf("chip8 version %s\n", version)
		return
	}

	// Ensure we have a ROM path for run/disasm
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: Missing ROM path\n\n")
		printUsage()
		os.Exit(1)
	}

	romPath := os.Args[2]

	// 2. Initialize CPU
	c := cpu.New()

	// 3. Command Execution
	switch command {
	case "run":
	if err := c.LoadROM(romPath); err != nil {
		fmt.Fprintf(os.Stderr, "Critical: Failed to load ROM: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully loaded: %s (%d bytes)\n", filepath.Base(romPath), c.ROMSize)
	fmt.Printf("Starting execution at PC: 0x%04X\n", c.PC)
	fmt.Println("--------------------------------------------------")

	c.DumpState()

	instructionsPerFrame := 10
	debug := true // toggle this off later

	for {
		// Run CPU (multiple instructions per frame)
		for i := 0; i < instructionsPerFrame; i++ {

			// Safety: prevent runaway PC
			if c.PC >= 4094 {
				fmt.Println("PC out of bounds, stopping")
				return
			}

			if !c.Step() {
				return
			}
		}

		// Debug output (disable for real ROMs)
		if debug {
			c.DumpDisplay()
			c.DumpState()
		}

		// 60 Hz timing (hardware tick)
		time.Sleep(time.Second / 60)

		// Delay Timer
		if c.DelayTimer > 0 {
			c.DelayTimer--
		}

		// Sound Timer
		if c.SoundTimer > 0 {
			c.SoundTimer--
			fmt.Println("BEEP") // placeholder for actual sound
		}
	}
	case "--disasm":
		if err := c.LoadROM(romPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Disassembling %s...\n\n", filepath.Base(romPath))
		// Starting address is 0x200 for standard CHIP-8
		c.Disassemble(0x200, 0x200+uint16(c.ROMSize))

	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n\n", command)
		printUsage()
		os.Exit(1)
	}
}