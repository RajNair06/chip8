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
	
	case 0x8000:
		switch n{
		case 0x0:
			fmt.Printf("LD V%X = V%X (%d)\n", x, y, c.V[y])
			c.V[x]=c.V[y]
		case 0x1:
			//OR
			fmt.Printf("OR V%X |= V%X (%d | %d)\n", x, y, c.V[x], c.V[y])
			c.V[x] |= c.V[y]
		case 0x2:
			fmt.Printf("AND V%X &= V%X (%d & %d)\n", x, y, c.V[x], c.V[y])
			c.V[x]&=c.V[y]
		case 0x3:
			fmt.Printf("XOR V%X ^= V%X (%d ^ %d)\n", x, y, c.V[x], c.V[y])
			c.V[x]^=c.V[y]
		case 0x4:
			sum:=uint16(c.V[x])+uint16(c.V[y])
			if sum>255{
			c.V[0xF]=1

			}else{
				c.V[0xF]=0
			}
			fmt.Printf("ADD V%X += V%X (%d + %d = %d , carry=%d) \n",x,y,c.V[x],c.V[y],sum&0xFF,c.V[0xF])
			c.V[x]=byte(sum & 0xFF)
		case 0x5:
			if c.V[x]>=c.V[y]{
				c.V[0xF]=1
			}else{
				c.V[0xF]=0

			}
			fmt.Printf("SUB V%X = %d - %d → %d (VF=%d)\n",x,c.V[x],c.V[y],c.V[x]-c.V[y],c.V[0xF])
			c.V[x]-=c.V[y]
		
		case 0x6:
			// SHR shift right

			c.V[0xF]=c.V[x]& 0x1 // save lsb to register
			fmt.Printf("SHR V%X (before=%d, LSB=%d)\n",x,c.V[x],c.V[0xF])
			c.V[x]>>=1
		
		case 0x7:
			if c.V[y]>=c.V[x]{
				c.V[0xF]=1

			}else{
				c.V[0xF]=0
			}
			fmt.Printf("SUBN V%X = V%X- V%X (%d - %d = %d)\n",x, y, x,c.V[y],c.V[x],c.V[y]-c.V[x])
			c.V[x]=c.V[y]-c.V[x]
		
		case 0xE:
			// SHL left shift

			c.V[0xF]=(c.V[x]& 0x80)>>7
			fmt.Printf("SHL V%X (before=%d, MSB=%d)\n",x,c.V[x],c.V[0xF])
			c.V[x]<<=1
		
		default:
			fmt.Println("unknown 8XY* opcode")
		
		}
	case 0x9000:
		if n==0{
			if c.V[x]!=c.V[y]{
				fmt.Printf("SNE V%X != V%X -> SKIP NEXT \n",x,y)
			}else{
				fmt.Printf("SNE V%X == V%X -> NO SKIP \n",x,y)
			}
		}
	case 0xA000:
		fmt.Printf("LD I = 0x%03X\n", nnn)
		c.I = nnn
	case 0xD000:
		xcord:=c.V[x]%64
		ycord:=c.V[y]%32
		height:=int(n)

		c.V[0xF]=0
		//outer loop to iterate through rows
		for row:=0;row<height;row++{
			spriteByte:=c.Memory[c.I+uint16(row)]
			// iterate through each 8 bits in byte
			for col:=0;col<8;col++{
				if(spriteByte&(0x80>>col))!=0{
					dispx:=(int(xcord)+col)%64
					dispy:=(int(ycord)+row)%32
					index:=dispx+(dispy*64)

					if c.Display[index]==1{
						c.V[0xF]=1
					}
					c.Display[index]^=1
					
				}

			}

		}
	case 0xF000:
	switch opcode & 0xF0FF {

	case 0xF015:
		fmt.Printf("LD DT = V%X (%d)\n", x, c.V[x])
		c.DelayTimer = c.V[x]

	case 0xF007:
		fmt.Printf("LD V%X = DT (%d)\n", x, c.DelayTimer)
		c.V[x] = c.DelayTimer

	case 0xF018:
		fmt.Printf("LD ST = V%X (%d)\n", x, c.V[x])
		c.SoundTimer = c.V[x]

	default:
		fmt.Printf("UNKNOWN F OPCODE: %04X\n", opcode)
	}

	default:
		fmt.Println("UNKNOWN OPCODE")
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