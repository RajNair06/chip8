package cpu

import (
	
	"math/rand"
	"os"
)

type Chip8 struct {
	Memory     [4096]uint8
	V          [16]uint8
	I          uint16
	PC         uint16
	Stack      [16]uint16
	SP         uint16
	DelayTimer uint8
	SoundTimer uint8
	Display    [64 * 32]uint8
	Keys       [16]bool
}

func New() *Chip8 {
	c := &Chip8{
		PC: 0x200,
	}
	// Load Fonts at 0x50
	fonts := []uint8{
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
	copy(c.Memory[0x50:], fonts)
	return c
}

func (c *Chip8) LoadROM(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	copy(c.Memory[0x200:], data)
	return nil
}

func (c *Chip8) Step() {
	opcode := uint16(c.Memory[c.PC])<<8 | uint16(c.Memory[c.PC+1])
	c.PC += 2
	c.Execute(opcode)
}

func (c *Chip8) Execute(opcode uint16) {
	x := (opcode & 0x0F00) >> 8
	y := (opcode & 0x00F0) >> 4
	n := opcode & 0x000F
	nn := uint8(opcode & 0x00FF)
	nnn := opcode & 0x0FFF

	switch opcode & 0xF000 {
	case 0x0000:
		if opcode == 0x00E0 {
			for i := range c.Display {
				c.Display[i] = 0
			}
		} else if opcode == 0x00EE {
			if c.SP > 0 {
				c.SP--
				c.PC = c.Stack[c.SP]
			}
		}
	case 0x1000:
		c.PC = nnn
	case 0x2000:
		if c.SP < 16 {
			c.Stack[c.SP] = c.PC
			c.SP++
			c.PC = nnn
		}
	case 0x3000:
		if c.V[x] == nn {
			c.PC += 2
		}
	case 0x4000:
		if c.V[x] != nn {
			c.PC += 2
		}
	case 0x5000:
		if c.V[x] == c.V[y] {
			c.PC += 2
		}
	case 0x6000:
		c.V[x] = nn
	case 0x7000:
		c.V[x] += nn
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
			res := uint16(c.V[x]) + uint16(c.V[y])
			c.V[x] = uint8(res & 0xFF)
			if res > 0xFF {
				c.V[0xF] = 1
			} else {
				c.V[0xF] = 0
			}
		case 0x5:
			var carry uint8 = 1
			if c.V[x] < c.V[y] {
				carry = 0
			}
			c.V[x] -= c.V[y]
			c.V[0xF] = carry
		case 0x6: // SHR (Modern)
			flag := c.V[x] & 0x1
			c.V[x] >>= 1
			c.V[0xF] = flag
		case 0x7:
			var carry uint8 = 1
			if c.V[y] < c.V[x] {
				carry = 0
			}
			c.V[x] = c.V[y] - c.V[x]
			c.V[0xF] = carry
		case 0xE: // SHL (Modern)
			flag := (c.V[x] & 0x80) >> 7
			c.V[x] <<= 1
			c.V[0xF] = flag
		}
	case 0x9000:
		if c.V[x] != c.V[y] {
			c.PC += 2
		}
	case 0xA000:
		c.I = nnn
	case 0xB000:
		c.PC = nnn + uint16(c.V[0])
	case 0xC000:
		c.V[x] = uint8(rand.Intn(256)) & nn
	case 0xD000:
		// Draw Opcode with Clipping
		xCoord := int(c.V[x]) % 64
		yCoord := int(c.V[y]) % 32
		c.V[0xF] = 0

		for row := 0; row < int(n); row++ {
			if yCoord+row >= 32 {
				break
			}
			spriteByte := c.Memory[c.I+uint16(row)]
			for col := 0; col < 8; col++ {
				if xCoord+col >= 64 {
					break
				}
				if (spriteByte & (0x80 >> col)) != 0 {
					idx := (yCoord+row)*64 + (xCoord+col)
					if c.Display[idx] == 1 {
						c.V[0xF] = 1
					}
					c.Display[idx] ^= 1
				}
			}
		}
	case 0xE000:
		key := c.V[x] & 0x0F
		if nn == 0x9E && c.Keys[key] {
			c.PC += 2
		} else if nn == 0xA1 && !c.Keys[key] {
			c.PC += 2
		}
	case 0xF000:
		switch nn {
		case 0x07:
			c.V[x] = c.DelayTimer
		case 0x0A: // Wait for Key
			pressed := false
			for i, k := range c.Keys {
				if k {
					c.V[x] = uint8(i)
					pressed = true
					break
				}
			}
			if !pressed {
				c.PC -= 2
			}
		case 0x15:
			c.DelayTimer = c.V[x]
		case 0x18:
			c.SoundTimer = c.V[x]
		case 0x1E:
			c.I += uint16(c.V[x])
		case 0x29:
			c.I = 0x50 + uint16(c.V[x]&0x0F)*5
		case 0x33:
			c.Memory[c.I] = c.V[x] / 100
			c.Memory[c.I+1] = (c.V[x] / 10) % 10
			c.Memory[c.I+2] = c.V[x] % 10
		case 0x55:
			for i := uint16(0); i <= x; i++ {
				c.Memory[c.I+i] = c.V[i]
			}
		case 0x65:
			for i := uint16(0); i <= x; i++ {
				c.V[i] = c.Memory[c.I+i]
			}
		}
	}
}