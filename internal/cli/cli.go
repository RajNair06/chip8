package cli

import (
	"fmt"
	"os"
	"path/filepath"
)

const Version = "1.0.0"

func PrintUsage() {
	executable := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "CHIP-8 Virtual Machine - Version %s\n\n", Version)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s <command> <rom_path>\n\n", executable)
	fmt.Fprintf(os.Stderr, "Commands:\n")
	fmt.Fprintf(os.Stderr, "  run        Initialize and execute the specified ROM\n")
	fmt.Fprintf(os.Stderr, "  --disasm   Disassemble the specified ROM into human-readable opcodes\n")
	fmt.Fprintf(os.Stderr, "  --version  Show version information\n")
}