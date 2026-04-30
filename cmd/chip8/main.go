package main

import (
	"chip8/internal/cli"
	"chip8/internal/cpu"
	"chip8/internal/gui"
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	cfg, err := cli.Parse(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
		cli.PrintUsage()
		os.Exit(1)
	}

	switch cfg.Command {
	case "version":
		cli.PrintVersion()
		return
	case "help":
		cli.PrintUsage()
		return
	case "disasm":
		runDisasm(cfg.RomPath)
		return
	case "run":
		runEmulator(cfg)
	default:
		cli.PrintUsage()
		os.Exit(1)
	}
}

// ── Disassembler mode ─────────────────────────────────────────────────────────
func runDisasm(romPath string) {
	data, err := os.ReadFile(romPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading ROM: %v\n", err)
		os.Exit(1)
	}
	lines := cpu.DisasmROM(data)
	for _, l := range lines {
		fmt.Println(l)
	}
}

// ── Emulator mode ─────────────────────────────────────────────────────────────
func runEmulator(cfg cli.Config) {
	// Validate ROM
	if _, err := os.Stat(cfg.RomPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: ROM not found: %s\n", cfg.RomPath)
		os.Exit(1)
	}

	// 1. Initialize CPU
	c := cpu.New()
	if err := c.LoadROM(cfg.RomPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading ROM: %v\n", err)
		os.Exit(1)
	}

	// 2. Done channel for goroutine cleanup
	done := make(chan struct{})

	// ── Goroutine: CPU at 500Hz ────────────────────────────────────────────
	// Each tick = 1 instruction. 500Hz ≈ 2ms per instruction.
	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(cfg.Speed))
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c.Mu.Lock()
				c.Step()
				c.Mu.Unlock()
			}
		}
	}()

	// ── Goroutine: Timers at 60Hz ──────────────────────────────────────────
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				c.Mu.Lock()
				if c.DelayTimer > 0 {
					c.DelayTimer--
				}
				if c.SoundTimer > 0 {
					c.SoundTimer--
				}
				c.Mu.Unlock()
			}
		}
	}()

	// ── Goroutine: Metrics reporter ────────────────────────────────────────
	go cpu.MetricsReporter(c, done)

	// 3. Initialize GUI
	g := gui.NewGui(c, cfg.RomPath, cfg.Debug)

	// 4. Window setup
	if cfg.Debug {
		ebiten.SetWindowSize(gui.TotalW, gui.GameH)
		ebiten.SetWindowTitle("CHIP-8 Systems Emulator — DEBUG MODE")
	} else {
		ebiten.SetWindowSize(gui.GameW, gui.GameH)
		ebiten.SetWindowTitle("CHIP-8 Systems Emulator")
	}
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	fmt.Printf("┌──────────────────────────────────────────┐\n")
	fmt.Printf("│ CHIP-8 Systems Emulator v%s             │\n", cli.Version)
	fmt.Printf("│ ROM:   %-34s│\n", cfg.RomPath)
	fmt.Printf("│ Speed: %-4d Hz  Debug: %-18v│\n", cfg.Speed, cfg.Debug)
	fmt.Printf("│ Keys:  1234/QWER/ASDF/ZXCV  ESC=pause   │\n")
	fmt.Printf("└──────────────────────────────────────────┘\n")

	// 5. Run (blocks until window closed)
	if err := ebiten.RunGame(g); err != nil {
		fmt.Fprintf(os.Stderr, "Emulator error: %v\n", err)
		close(done)
		os.Exit(1)
	}

	// Cleanup goroutines
	close(done)
}