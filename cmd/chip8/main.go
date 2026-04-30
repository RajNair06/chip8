package main

import (
	"chip8/internal/cpu"
	"chip8/internal/gui"
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	// Expecting: go run ./cmd/chip8/main.go run path/to/rom.ch8
	// os.Args[0] = program path
	// os.Args[1] = "run"
	// os.Args[2] = ROM path
	if len(os.Args) < 3 {
		fmt.Println("Usage: chip8 run <rom_path>")
		os.Exit(1)
	}

	romPath := os.Args[2]

	// Check if file exists before initializing everything
	if _, err := os.Stat(romPath); os.IsNotExist(err) {
		fmt.Printf("Error: ROM file not found at %s\n", romPath)
		os.Exit(1)
	}

	// 1. Initialize the Hardware
	c := cpu.New()

	// 2. Load the ROM
	if err := c.LoadROM(romPath); err != nil {
		fmt.Printf("Error loading ROM: %v\n", err)
		os.Exit(1)
	}

	// 3. Initialize the GUI (Ebiten Game implementation)
	g := gui.NewGui(c)

	// 4. Window Configuration
	// Window size is 15x the base CHIP-8 resolution (960x480)
	ebiten.SetWindowSize(960, 480)
	ebiten.SetWindowTitle("CHIP-8 Systems Emulator")
	
	// Ensures the buffer is clean every frame to prevent 
	// pixel ghosting and stretching artifacts.
	ebiten.SetScreenClearedEveryFrame(true)

	// 5. Fire up the VM
	fmt.Printf("Starting emulation: %s\n", romPath)
	if err := ebiten.RunGame(g); err != nil {
		fmt.Printf("Emulator crashed: %v\n", err)
		panic(err)
	}
}