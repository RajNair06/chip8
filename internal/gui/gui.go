package gui

import (
	"chip8/internal/cpu"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	ScreenWidth  = 64
	ScreenHeight = 32
	Upscale      = 15
)

var keyMap = map[ebiten.Key]int{
	ebiten.Key1: 0x1, ebiten.Key2: 0x2, ebiten.Key3: 0x3, ebiten.Key4: 0xC,
	ebiten.KeyQ: 0x4, ebiten.KeyW: 0x5, ebiten.KeyE: 0x6, ebiten.KeyR: 0xD,
	ebiten.KeyA: 0x7, ebiten.KeyS: 0x8, ebiten.KeyD: 0x9, ebiten.KeyF: 0xE,
	ebiten.KeyZ: 0xA, ebiten.KeyX: 0x0, ebiten.KeyC: 0xB, ebiten.KeyV: 0xF,
}

var (
	pixelOn  = color.RGBA{0, 255, 65, 255}  // Matrix Green
	pixelOff = color.RGBA{20, 20, 20, 255}   // Dark Grey
)

type Gui struct {
	CPU        *cpu.Chip8
	offscreen  *ebiten.Image // 64x32 internal buffer
}

func NewGui(c *cpu.Chip8) *Gui {
	return &Gui{
		CPU:       c,
		offscreen: ebiten.NewImage(ScreenWidth, ScreenHeight),
	}
}

func (g *Gui) Update() error {
	// 1. Key Polling: Must happen before stepping
	for ebKey, chipKey := range keyMap {
		g.CPU.Keys[chipKey] = ebiten.IsKeyPressed(ebKey)
	}

	// 2. CPU Cycles: 15 cycles/frame * 60 FPS = 900Hz
	for i := 0; i < 15; i++ {
		g.CPU.Step()
	}

	// 3. Timers: Decrement at 60Hz
	if g.CPU.DelayTimer > 0 {
		g.CPU.DelayTimer--
	}
	if g.CPU.SoundTimer > 0 {
		g.CPU.SoundTimer--
	}

	return nil
}

func (g *Gui) Draw(screen *ebiten.Image) {
	// Build a raw RGBA image from the CHIP-8 display buffer
	img := image.NewRGBA(image.Rect(0, 0, ScreenWidth, ScreenHeight))
	for i, pixel := range g.CPU.Display {
		x := i % ScreenWidth
		y := i / ScreenWidth
		if pixel == 1 {
			img.SetRGBA(x, y, pixelOn)
		} else {
			img.SetRGBA(x, y, pixelOff)
		}
	}

	// Upload pixel data to the offscreen GPU texture
	g.offscreen.WritePixels(img.Pix)

	// Scale the 64x32 texture up to the full 960x480 window
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(Upscale), float64(Upscale))
	screen.DrawImage(g.offscreen, op)
}

// Layout returns the LOGICAL (window) resolution, not the CHIP-8 resolution.
// Ebiten uses this to know how big the screen surface is.
// We handle all scaling ourselves in Draw().
func (g *Gui) Layout(outsideWidth, outsideHeight int) (int, int) {
	return ScreenWidth * Upscale, ScreenHeight * Upscale
}