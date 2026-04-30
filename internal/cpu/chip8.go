package cpu

import (
	"fmt"
	"math/rand"
	"os"
)

type Chip8 struct{
	Memory [4096]byte
	ROMSize int
	V [16]byte
	I uint16
	PC uint16

	Stack [16]uint16
	SP uint8

	DelayTimer byte
	SoundTimer byte

	Display [64*32]byte
	Keys[16]bool



}

var fontSet = []byte{
    0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
    0x20, 0x60, 0x20, 0x20, 0x70, // 1
    0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
    0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
    0x90, 0x90, 0xF0, 0x10, 0x10, // 4
    0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
    0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
    0xF0, 0x10, 0x20, 0x40, 0x40, // 7
    0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
    0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
    0xF0, 0x90, 0xF0, 0x90, 0x90, // A
    0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
    0xF0, 0x80, 0x80, 0x80, 0xF0, // C
    0xE0, 0x90, 0x90, 0x90, 0xE0, // D
    0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
    0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func New() *Chip8{
	c:=&Chip8{}
	c.PC=0x200
	copy(c.Memory[:],fontSet)
	
	return  c

}

func (c *Chip8) LoadROM(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to read ROM: %w", err)
    }

    // Safety Check: CHIP-8 memory ends at 0xFFF (4095)
    // 0xFFF - 0x200 = 3583 bytes available for ROM
    if len(data) > (4096 - 0x200) {
        return fmt.Errorf("ROM file too large (%d bytes)", len(data))
    }

    c.ROMSize = len(data)

    // Load data into memory starting at 0x200
    for i := 0; i < len(data); i++ {
        c.Memory[0x200+i] = data[i]
    }

    return nil
}
func (c *Chip8) DumpState() {
	
	for i := 0; i < 16; i++ {
		println("V", i, ":", c.V[i])
		fmt.Printf("DT: %d | ST: %d\n", c.DelayTimer, c.SoundTimer)
	}
}

func (c *Chip8) printInstruction(opcode uint16) {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := opcode & 0x00FF
	nnn := opcode & 0x0FFF

	switch opcode & 0xF000 {
	case 0x0000:
		switch opcode {
		case 0x00E0:
			fmt.Println("CLS")
		case 0x00EE:
			fmt.Println("RET")
		default:
			fmt.Printf("SYS 0x%X\n", nnn)
		}
	case 0x1000:
    	fmt.Printf("JP 0x%03X\n", nnn)
	case 0x6000:
		fmt.Printf("LD V%X, 0x%02X\n", x, nn)
	case 0x7000:
		fmt.Printf("ADD V%X, 0x%02X\n", x, nn)

	case 0x8000:
		switch opcode & 0x000F {
		case 0x0:
			fmt.Printf("LD V%X, V%X\n", x, y)
		case 0x1:
			fmt.Printf("OR V%X, V%X\n", x, y)
		case 0x2:
			fmt.Printf("AND V%X, V%X\n", x, y)
		case 0x3:
			fmt.Printf("XOR V%X, V%X\n", x, y)
		case 0x4:
			fmt.Printf("ADD V%X, V%X\n", x, y)
		case 0x5:
			fmt.Printf("SUB V%X, V%X\n", x, y)
		case 0x6:
			fmt.Printf("SHR V%X {, V%X}\n", x, y)
		case 0x7:
			fmt.Printf("SUBN V%X, V%X\n", x, y)
		case 0xE:
			fmt.Printf("SHL V%X {, V%X}\n", x, y)
		default:
			fmt.Printf("UNKNOWN 8-series: %04X\n", opcode)
		}

	case 0x9000:
		if n == 0 {
			fmt.Printf("SNE V%X, V%X\n", x, y)
		} else {
			fmt.Printf("UNKNOWN 9-series: %04X\n", opcode)
		}

	case 0xA000:
		fmt.Printf("LD I, 0x%03X\n", nnn)

	case 0xD000:
		fmt.Printf("DRW V%X, V%X, %X\n", x, y, n)
	case 0xF000:
	switch opcode & 0xF0FF {

	case 0xF015:
		fmt.Printf("LD DT, V%X\n", x)

	case 0xF007:
		fmt.Printf("LD V%X, DT\n", x)

	case 0xF018:
		fmt.Printf("LD ST, V%X\n", x)

	default:
		fmt.Printf("UNKNOWN F-series: %04X\n", opcode)
	}
	default:
		fmt.Printf("UNKNOWN OPCODE: %04X\n", opcode)
	}
}

func (c *Chip8) Disassemble(start, end uint16){
	pc:=start
	for pc<end{
		opcode:=uint16(c.Memory[pc])<<8 | uint16(c.Memory[pc+1])
		fmt.Printf("%04X: %04X ",pc,opcode)
		c.printInstruction(opcode)
		pc+=2
	}
}

func (c *Chip8) Step() bool {
	// 1. Safety check: Ensure we are within memory bounds
	if c.PC >= 4094 {
		return false
	}

	// 2. Fetch: Get the 16-bit opcode from two 8-bit memory locations
	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])

	// 3. Increment PC: Point to the NEXT instruction BEFORE executing the current one
	// This is important because instructions like Skips (SE/SNE) will increment it again (+2)
	c.PC += 2

	// 4. Decode & Execute
	c.Execute(opcode)

	// 5. Return true to keep the loop running
	return true
}

func (c *Chip8) Execute(opcode uint16) {
    x := (opcode & 0x0F00) >> 8
    y := (opcode & 0x00F0) >> 4
    n := opcode & 0x000F
    nn := byte(opcode & 0x00FF)
    nnn := opcode & 0x0FFF

    switch opcode & 0xF000 {

    case 0x0000:
        switch opcode {
        case 0x00E0:
            fmt.Println("CLS → clear display")
            for i := range c.Display {
                c.Display[i] = 0
            }
        case 0x00EE:
            if c.SP == 0 {
                fmt.Println("ERROR: stack underflow on RET")
                return
            }
            c.SP--
            retAddr := c.Stack[c.SP]
            fmt.Printf("RET → pop PC (0x%03X) from stack\n", retAddr)
            c.PC = retAddr
        default:
            fmt.Println("SYS (ignored)")
        }

    case 0x1000:
        fmt.Printf("JP → jump to 0x%03X\n", nnn)
        c.PC = nnn

    case 0x2000:
        fmt.Printf("CALL 0x%03X → push PC (0x%03X) to stack\n", nnn, c.PC)
        c.Stack[c.SP] = c.PC
        c.SP++
        c.PC = nnn

    case 0x3000:
        if c.V[x] == nn {
            fmt.Printf("SE V%X == %d → skip next\n", x, nn)
            c.PC += 2
        }

    case 0x4000:
        if c.V[x] != nn {
            fmt.Printf("SNE V%X != %d → skip next\n", x, nn)
            c.PC += 2
        }

    case 0x5000:
        if n == 0 && c.V[x] == c.V[y] {
            fmt.Printf("SE V%X == V%X → skip next\n", x, y)
            c.PC += 2
        }

    case 0x6000:
        fmt.Printf("LD V%X = %d\n", x, nn)
        c.V[x] = nn

    case 0x7000:
        c.V[x] += nn
        fmt.Printf("ADD V%X += %d\n", x, nn)

    case 0x8000:
        switch n {
        case 0x0:
            c.V[x] = c.V[y]
        case 0x1:
            c.V[x] |= c.V[y]
        case 0x2:
            c.V[x] &= c.V[y]
        case 0x3:
            c.V[x] ^= c.V[y]
        case 0x4:
            sum := uint16(c.V[x]) + uint16(c.V[y])
            c.V[0xF] = 0
            if sum > 255 { c.V[0xF] = 1 }
            c.V[x] = byte(sum & 0xFF)
        case 0x5:
            res := c.V[x] >= c.V[y]
            c.V[x] -= c.V[y]
            if res { c.V[0xF] = 1 } else { c.V[0xF] = 0 }
        case 0x6:
            c.V[0xF] = c.V[x] & 0x1
            c.V[x] >>= 1
        case 0x7:
            res := c.V[y] >= c.V[x]
            c.V[x] = c.V[y] - c.V[x]
            if res { c.V[0xF] = 1 } else { c.V[0xF] = 0 }
        case 0xE:
            c.V[0xF] = (c.V[x] & 0x80) >> 7
            c.V[x] <<= 1
        }

    case 0x9000:
        if n == 0 && c.V[x] != c.V[y] {
            fmt.Printf("SNE V%X != V%X -> SKIP\n", x, y)
            c.PC += 2
        }

    case 0xA000:
        c.I = nnn

    case 0xB000:
        c.PC = nnn + uint16(c.V[0])

    case 0xC000:
        c.V[x] = byte(rand.Intn(256)) & nn

    case 0xD000:
        // Sprite Drawing with Modulo Wrapping to prevent "Barcoding"
        xBase := int(c.V[x])
        yBase := int(c.V[y])
        c.V[0xF] = 0

        for row := 0; row < int(n); row++ {
            yCoord := (yBase + row) % 32
            spriteByte := c.Memory[c.I+uint16(row)]

            for col := 0; col < 8; col++ {
                xCoord := (xBase + col) % 64
                if (spriteByte & (0x80 >> col)) != 0 {
                    index := xCoord + (yCoord * 64)
                    if c.Display[index] == 1 {
                        c.V[0xF] = 1
                    }
                    c.Display[index] ^= 1
                }
            }
			fmt.Println("--- SCREEN DUMP ---")
for y := 0; y < 32; y++ {
    for x := 0; x < 64; x++ {
        if c.Display[x+(y*64)] == 1 {
            fmt.Print("#")
        } else {
            fmt.Print(".")
        }
    }
    fmt.Println()
}
        }

    case 0xE000:
        switch nn {
        case 0x9E:
            if c.Keys[c.V[x]] { c.PC += 2 }
        case 0xA1:
            if !c.Keys[c.V[x]] { c.PC += 2 }
        }

    case 0xF000:
        switch nn {
        case 0x07:
            c.V[x] = c.DelayTimer
        case 0x0A:
            keyState := false
            for i, pressed := range c.Keys {
                if pressed {
                    c.V[x] = byte(i)
                    keyState = true
                    break
                }
            }
            if !keyState { c.PC -= 2 } // Keep repeating this instruction
        case 0x15:
            c.DelayTimer = c.V[x]
        case 0x18:
            c.SoundTimer = c.V[x]
        case 0x1E:
            c.I += uint16(c.V[x])
        case 0x29:
            c.I = uint16(c.V[x]) * 5
        case 0x33:
            c.Memory[c.I] = c.V[x] / 100
            c.Memory[c.I+1] = (c.V[x] / 10) % 10
            c.Memory[c.I+2] = c.V[x] % 10
        case 0x55:
            for i := 0; i <= int(x); i++ {
                c.Memory[int(c.I)+i] = c.V[i]
            }
        case 0x65:
            for i := 0; i <= int(x); i++ {
                c.V[i] = c.Memory[int(c.I)+i]
            }
        }

    default:
        fmt.Printf("UNKNOWN OPCODE: %04X\n", opcode)
    }
}

func (c *Chip8) DumpDisplay() {
    for y := 0; y < 32; y++ {
        for x := 0; x < 64; x++ {
            if c.Display[y*64+x] == 1 {
                fmt.Print("█")
            } else {
                fmt.Print(" ")
            }
        }
        fmt.Println()
    }
}