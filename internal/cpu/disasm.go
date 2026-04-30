package cpu

import "fmt"

// Disassemble decodes a single opcode at addr into a human-readable string.
// Used by both the CLI disassembler and the live debug panel.
func Disassemble(addr uint16, opcode uint16) string {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := opcode & 0x00FF
	nnn := opcode & 0x0FFF

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0:
			return "CLS"
		case 0x00EE:
			return "RET"
		default:
			return fmt.Sprintf("SYS   0x%03X", nnn)
		}
	case 0x1000:
		return fmt.Sprintf("JP    0x%03X", nnn)
	case 0x2000:
		return fmt.Sprintf("CALL  0x%03X", nnn)
	case 0x3000:
		return fmt.Sprintf("SE    V%X, 0x%02X", x, nn)
	case 0x4000:
		return fmt.Sprintf("SNE   V%X, 0x%02X", x, nn)
	case 0x5000:
		return fmt.Sprintf("SE    V%X, V%X", x, y)
	case 0x6000:
		return fmt.Sprintf("LD    V%X, 0x%02X", x, nn)
	case 0x7000:
		return fmt.Sprintf("ADD   V%X, 0x%02X", x, nn)
	case 0x8000:
		switch n {
		case 0x0:
			return fmt.Sprintf("LD    V%X, V%X", x, y)
		case 0x1:
			return fmt.Sprintf("OR    V%X, V%X", x, y)
		case 0x2:
			return fmt.Sprintf("AND   V%X, V%X", x, y)
		case 0x3:
			return fmt.Sprintf("XOR   V%X, V%X", x, y)
		case 0x4:
			return fmt.Sprintf("ADD   V%X, V%X", x, y)
		case 0x5:
			return fmt.Sprintf("SUB   V%X, V%X", x, y)
		case 0x6:
			return fmt.Sprintf("SHR   V%X", x)
		case 0x7:
			return fmt.Sprintf("SUBN  V%X, V%X", x, y)
		case 0xE:
			return fmt.Sprintf("SHL   V%X", x)
		}
	case 0x9000:
		return fmt.Sprintf("SNE   V%X, V%X", x, y)
	case 0xA000:
		return fmt.Sprintf("LD    I, 0x%03X", nnn)
	case 0xB000:
		return fmt.Sprintf("JP    V0, 0x%03X", nnn)
	case 0xC000:
		return fmt.Sprintf("RND   V%X, 0x%02X", x, nn)
	case 0xD000:
		return fmt.Sprintf("DRW   V%X, V%X, %d", x, y, n)
	case 0xE000:
		switch nn {
		case 0x9E:
			return fmt.Sprintf("SKP   V%X", x)
		case 0xA1:
			return fmt.Sprintf("SKNP  V%X", x)
		}
	case 0xF000:
		switch nn {
		case 0x07:
			return fmt.Sprintf("LD    V%X, DT", x)
		case 0x0A:
			return fmt.Sprintf("LD    V%X, K", x)
		case 0x15:
			return fmt.Sprintf("LD    DT, V%X", x)
		case 0x18:
			return fmt.Sprintf("LD    ST, V%X", x)
		case 0x1E:
			return fmt.Sprintf("ADD   I, V%X", x)
		case 0x29:
			return fmt.Sprintf("LD    F, V%X", x)
		case 0x33:
			return fmt.Sprintf("LD    B, V%X", x)
		case 0x55:
			return fmt.Sprintf("LD    [I], V%X", x)
		case 0x65:
			return fmt.Sprintf("LD    V%X, [I]", x)
		}
	}
	return fmt.Sprintf("DATA  0x%04X", opcode)
}

// DisasmROM disassembles an entire ROM file and returns annotated lines.
// It performs a two-pass scan: first collects jump targets, then annotates.
func DisasmROM(romData []byte) []string {
	// --- Pass 1: collect all jump/call targets for label generation ---
	targets := map[uint16]bool{}
	for i := 0; i+1 < len(romData); i += 2 {
		op := uint16(romData[i])<<8 | uint16(romData[i+1])
		nnn := op & 0x0FFF
		switch op & 0xF000 {
		case 0x1000, 0x2000, 0xB000:
			targets[nnn] = true
		}
	}

	// --- Pass 2: emit annotated disassembly ---
	var lines []string
	lines = append(lines, fmt.Sprintf("%-6s  %-4s  %-20s  %s", "ADDR", "HEX", "INSTRUCTION", "NOTE"))
	lines = append(lines, "------  ----  --------------------  --------")

	for i := 0; i+1 < len(romData); i += 2 {
		addr := uint16(0x200 + i)
		op := uint16(romData[i])<<8 | uint16(romData[i+1])
		instr := Disassemble(addr, op)

		note := ""
		nnn := op & 0x0FFF
		switch op & 0xF000 {
		case 0x1000:
			if nnn == addr {
				note = "; <-- infinite loop / halt"
			} else {
				note = fmt.Sprintf("; --> 0x%03X", nnn)
			}
		case 0x2000:
			note = fmt.Sprintf("; call 0x%03X", nnn)
		case 0x00EE & 0xF000:
			if op == 0x00EE {
				note = "; return from subroutine"
			}
		case 0xD000:
			n := op & 0x000F
			note = fmt.Sprintf("; draw %d rows", n)
		case 0xF000:
			switch op & 0x00FF {
			case 0x0A:
				note = "; BLOCKING: wait for keypress"
			case 0x33:
				note = "; BCD decode to I,I+1,I+2"
			}
		}

		label := "      "
		if targets[addr] {
			label = fmt.Sprintf("L%04X:", addr)
		}

		lines = append(lines, fmt.Sprintf("%s  %04X  %-20s  %s", label, op, instr, note))
	}
	return lines
}