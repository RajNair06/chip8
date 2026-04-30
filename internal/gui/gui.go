package gui

import (
	"chip8/internal/cpu"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	ScreenWidth =64
	ScreenHeight=32
	Upscale=15 // Each CHIP-8 pixel becomes 15x15 on  monitor
)

type Gui struct {
	CPU *cpu.Chip8
}
func (g *Gui) Update() error {
    // Run multiple CPU cycles per frame to reach ~600Hz
    // 10 cycles * 60 frames per second = 600 instructions per second
    cyclesPerFrame := 10 

    for i := 0; i < cyclesPerFrame; i++ {
        g.CPU.Step()
    }

    // Update timers at 60Hz (once per Update call)
    if g.CPU.DelayTimer > 0 {
        g.CPU.DelayTimer--
    }
    if g.CPU.SoundTimer > 0 {
        g.CPU.SoundTimer--
    }

    return nil
}
func (g *Gui) Draw(screen *ebiten.Image) {
    screen.Fill(color.RGBA{20, 20, 20, 255})

    // REMOVE THIS LINE: vector.DrawFilledRect(screen, 0, 0, ... red color ...)

    for y := 0; y < ScreenHeight; y++ {
        for x := 0; x < ScreenWidth; x++ {
            if g.CPU.Display[y*ScreenWidth+x] == 1 {
                vector.DrawFilledRect(
                    screen,
                    float32(x*Upscale),
                    float32(y*Upscale),
                    float32(Upscale),
                    float32(Upscale),
                    color.RGBA{0, 255, 65, 255},
                    false,
                )
            }
        }
    }
}

func (g *Gui) Layout(outsideWidth,outsideHeight int)(int,int){
	return  ScreenWidth *Upscale,ScreenHeight*Upscale
}