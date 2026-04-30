package main

import (
	"fmt"
	"os"
	"path/filepath"
	

	"chip8/internal/cli"
	"chip8/internal/cpu"
	"chip8/internal/gui"

	"github.com/hajimehoshi/ebiten/v2"
)

const version = "1.0.0"

func main() {
	// 1. Basic Argument Validation
	if len(os.Args) < 2 {
		cli.PrintUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle version flag early
	if command == "--version" || command == "-v" {
		fmt.Printf("chip8 version %s\n", cli.Version)
		return
	}

	// Ensure we have a ROM path for run/disasm
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Error: Missing ROM path\n\n")
		cli.PrintUsage()
		os.Exit(1)
	}

	romPath := os.Args[2]

	// 2. Initialize CPU
	c := cpu.New()

	// 3. Command Execution
	switch command {
	case "run":
        // 1. Load the ROM into memory
        if err := c.LoadROM(romPath); err != nil {
            fmt.Fprintf(os.Stderr, "Critical: Failed to load ROM: %v\n", err)
            os.Exit(1)
        }

        fmt.Printf("Successfully loaded: %s (%d bytes)\n", filepath.Base(romPath), c.ROMSize)
        fmt.Printf("Entry Point: 0x%04X\n", c.PC)
        fmt.Println("--------------------------------------------------")

        // 2. Initialize the GUI bridge
        ui := &gui.Gui{
            CPU: c,
        }

        // 3. Configure the Window
        // ScreenWidth (64) * Upscale (15) = 960px wide
        // ScreenHeight (32) * Upscale (15) = 480px high
        ebiten.SetWindowSize(960, 480) 
        ebiten.SetWindowTitle("CHIP-8 Emulator | " + filepath.Base(romPath))
        
        // We set TPS to 60 because that's when we'll decrement timers
        ebiten.SetTPS(60)
        ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

        // 4. Start the Engine
        // This is a blocking call. Ebiten takes over the loop.
        if err := ebiten.RunGame(ui); err != nil {
            fmt.Fprintf(os.Stderr, "Emulator Crash: %v\n", err)
            os.Exit(1)
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
		cli.PrintUsage()
		os.Exit(1)
	}
}