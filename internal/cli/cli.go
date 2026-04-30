package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const Version = "2.0.0"

type Config struct {
	Command string // "run" | "disasm" | "version"
	RomPath string
	Debug   bool   // --debug flag: show debug panel
	Speed   int    // CPU Hz (default 500)
}

func Parse(args []string) (Config, error) {
	cfg := Config{Speed: 500}

	if len(args) < 2 {
		return cfg, fmt.Errorf("no command given")
	}

	// Scan flags anywhere in args
	remaining := []string{}
	for _, a := range args[1:] {
		switch a {
		case "--debug", "-d":
			cfg.Debug = true
		case "--version", "-v":
			cfg.Command = "version"
			return cfg, nil
		case "--help", "-h":
			cfg.Command = "help"
			return cfg, nil
		default:
			remaining = append(remaining, a)
		}
	}

	if len(remaining) == 0 {
		return cfg, fmt.Errorf("no command given")
	}

	switch remaining[0] {
	case "run":
		cfg.Command = "run"
		if len(remaining) < 2 {
			return cfg, fmt.Errorf("'run' requires a ROM path")
		}
		cfg.RomPath = remaining[1]
	case "--disasm", "disasm":
		cfg.Command = "disasm"
		if len(remaining) < 2 {
			return cfg, fmt.Errorf("'--disasm' requires a ROM path")
		}
		cfg.RomPath = remaining[1]
	default:
		return cfg, fmt.Errorf("unknown command: %q", remaining[0])
	}

	return cfg, nil
}

func PrintUsage() {
	exe := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "\n╔══════════════════════════════════════╗\n")
	fmt.Fprintf(os.Stderr, "║   CHIP-8 Systems Emulator  v%-8s║\n", Version)
	fmt.Fprintf(os.Stderr, "╚══════════════════════════════════════╝\n\n")
	fmt.Fprintf(os.Stderr, "USAGE:\n")
	fmt.Fprintf(os.Stderr, "  %s <command> [flags] <rom>\n\n", exe)
	fmt.Fprintf(os.Stderr, "COMMANDS:\n")
	fmt.Fprintf(os.Stderr, "  run       <rom>   Run a CHIP-8 ROM\n")
	fmt.Fprintf(os.Stderr, "  disasm    <rom>   Disassemble ROM to stdout\n")
	fmt.Fprintf(os.Stderr, "  --version         Show version\n\n")
	fmt.Fprintf(os.Stderr, "FLAGS:\n")
	fmt.Fprintf(os.Stderr, "  --debug           Show live debug panel (registers, memory viewer)\n\n")
	fmt.Fprintf(os.Stderr, "KEYBOARD MAP:\n")
	fmt.Fprintf(os.Stderr, "  CHIP-8   →  PC\n")
	fmt.Fprintf(os.Stderr, "  1 2 3 C  →  1 2 3 4\n")
	fmt.Fprintf(os.Stderr, "  4 5 6 D  →  Q W E R\n")
	fmt.Fprintf(os.Stderr, "  7 8 9 E  →  A S D F\n")
	fmt.Fprintf(os.Stderr, "  A 0 B F  →  Z X C V\n\n")
	fmt.Fprintf(os.Stderr, "  ESC      →  Pause/Resume\n\n")
	fmt.Fprintf(os.Stderr, "EXAMPLES:\n")
	fmt.Fprintf(os.Stderr, "  %s run roms/invaders.ch8\n", exe)
	fmt.Fprintf(os.Stderr, "  %s run --debug roms/maze.ch8\n", exe)
	fmt.Fprintf(os.Stderr, "  %s disasm roms/ibm.ch8\n\n", exe)
}

func PrintVersion() {
	fmt.Printf("CHIP-8 Systems Emulator v%s\n", Version)
	fmt.Printf("Phases complete: 1–10 (Professional Grade)\n")
	fmt.Printf("Opcodes: 35/35 ✓\n")
	fmt.Printf("Features: Goroutine CPU, 500Hz timing, Debug Panel, Disassembler\n")
}