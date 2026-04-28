package cpu

import (
	"fmt"
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

func New() *Chip8{
	c:=&Chip8{}
	c.PC=0x200
	return  c

}

func (c *Chip8) LoadROM(path string)error{
	data,err:=os.ReadFile(path)
	if err!=nil{
		return err
	}
	c.ROMSize=len(data)
	for i:=0;i<len(data);i++{
		c.Memory[0x200+i]=data[i]
	}
	return nil 
}
func (c *Chip8) DumpState() {
	
	for i := 0; i < 16; i++ {
		println("V", i, ":", c.V[i])
	}
}

func (c *Chip8) printInstruction(opcode uint16){
		x:=(opcode & 0x0F00)>>8
		y:=(opcode & 0x00F0)>>4
		n:=opcode & 0x000F
		nn:=opcode & 0x00FF
		nnn:=opcode & 0x0FFF

	switch opcode & 0xF000 {
	case 0x6000: 
	fmt.Printf("LD V%X, 0x%X\n", x, nn) 
	case 0x7000: 
	fmt.Printf("ADD V%X, 0x%X\n", x, nn) 
	case 0xA000: 
	fmt.Printf("LD I, 0x%X\n", nnn) 
	case 0xD000: fmt.Printf("DRW V%X, V%X, %X\n", x, y, n) 
	case 0x0000:
	switch opcode {
	case 0x00E0:
		fmt.Println("CLS")
	case 0x00EE:
		fmt.Println("RET")
	default:
		fmt.Println("SYS (ignored)")
	}
	default:
		fmt.Println("SYS (ignored)")
	}
}

func (c *Chip8) Disassemble(start , end uint16){
	pc:=start
	for pc<end{
		opcode:=uint16(c.Memory[pc])<<8 | uint16(c.Memory[pc+1])
		fmt.Printf("%04X: %04X ",pc,opcode)
		c.printInstruction(opcode)
		pc+=2
	}
}

func (c *Chip8) Step() bool{
	if c.PC >= 0x200+uint16(c.ROMSize) {
		fmt.Println("Reached end of ROM, stopping")
		return false
	}
	opcode:=uint16(c.Memory[c.PC])<<8|uint16(c.Memory[c.PC+1])
	
	fmt.Printf("PC: %04X | OPCODE: %04X\n", c.PC, opcode)
	c.PC+=2
	// execute instrution
	c.Execute(opcode)
	return true

}


func (c *Chip8) Execute(opcode uint16){
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
		} else {
			fmt.Printf("SE V%X != %d → no skip\n", x, nn)
		}

	case 0x4000:
		if c.V[x] != nn {
			fmt.Printf("SNE V%X != %d → skip next\n", x, nn)
			c.PC += 2
		} else {
			fmt.Printf("SNE V%X == %d → no skip\n", x, nn)
		}

	case 0x5000:
		if n == 0 && c.V[x] == c.V[y] {
			fmt.Printf("SE V%X == V%X → skip next\n", x, y)
			c.PC += 2
		} else {
			fmt.Printf("SE V%X != V%X → no skip\n", x, y)
		}

	case 0x6000:
		fmt.Printf("LD V%X = %d\n", x, nn)
		c.V[x] = nn

	case 0x7000:
		before := c.V[x]
		c.V[x] += nn
		fmt.Printf("ADD V%X: %d + %d = %d\n", x, before, nn, c.V[x])

	default:
		fmt.Println("UNKNOWN OPCODE")
	}
}
