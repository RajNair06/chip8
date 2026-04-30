package gui

import (
	"chip8/internal/cpu"
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ── Layout constants ──────────────────────────────────────────────────────────
const (
	ChipW   = 64
	ChipH   = 32
	Upscale = 15

	GameW = ChipW * Upscale // 960
	GameH = ChipH * Upscale // 480

	PanelW = 340           // debug panel width
	TotalW = GameW + PanelW // 1300

	MemViewRows  = 16 // rows visible in memory viewer
	MemBytesRow  = 8  // bytes per row in memory viewer
)

// ── Color palette ─────────────────────────────────────────────────────────────
var (
	colBg        = color.RGBA{10, 10, 14, 255}
	colPixelOn   = color.RGBA{0, 255, 100, 255}   // matrix green
	colPixelOff  = color.RGBA{15, 22, 15, 255}    // very dark green tint
	colPanelBg   = color.RGBA{14, 16, 20, 255}
	colPanelBdr  = color.RGBA{0, 180, 70, 255}    // green border
	colHdr       = color.RGBA{0, 255, 100, 255}
	colLabel     = color.RGBA{100, 140, 110, 255}
	colValue     = color.RGBA{220, 255, 230, 255}
	colHighlight = color.RGBA{255, 220, 0, 255}   // yellow for PC
	colMemByte   = color.RGBA{140, 180, 155, 255}
	colMemPC     = color.RGBA{0, 255, 100, 255}
	colMemAddr   = color.RGBA{60, 100, 75, 255}
	colDim       = color.RGBA{60, 80, 65, 255}
	colIPS       = color.RGBA{0, 200, 255, 255}
)

// ── Key map ───────────────────────────────────────────────────────────────────
var keyMap = map[ebiten.Key]int{
	ebiten.Key1: 0x1, ebiten.Key2: 0x2, ebiten.Key3: 0x3, ebiten.Key4: 0xC,
	ebiten.KeyQ: 0x4, ebiten.KeyW: 0x5, ebiten.KeyE: 0x6, ebiten.KeyR: 0xD,
	ebiten.KeyA: 0x7, ebiten.KeyS: 0x8, ebiten.KeyD: 0x9, ebiten.KeyF: 0xE,
	ebiten.KeyZ: 0xA, ebiten.KeyX: 0x0, ebiten.KeyC: 0xB, ebiten.KeyV: 0xF,
}

// ── Gui ───────────────────────────────────────────────────────────────────────
type Gui struct {
	CPU      *cpu.Chip8
	DebugOn  bool
	RomPath  string

	offscreen *ebiten.Image

	// Memory viewer scroll
	memScroll    int // first row index
	memScrollMax int

	// Frame timing
	lastFrame   time.Time
	frameJitter float64

	// Paused state
	paused bool
}

func NewGui(c *cpu.Chip8, romPath string, debug bool) *Gui {
	g := &Gui{
		CPU:          c,
		DebugOn:      debug,
		RomPath:      romPath,
		offscreen:    ebiten.NewImage(ChipW, ChipH),
		memScrollMax: (4096/MemBytesRow - MemViewRows),
	}
	return g
}

// ── Update (runs every frame at 60Hz) ─────────────────────────────────────────
func (g *Gui) Update() error {
	// Pause toggle
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		g.paused = !g.paused
	}
	// Memory viewer scroll
	if g.DebugOn {
		_, dy := ebiten.Wheel()
		if dy > 0 {
			g.memScroll--
		} else if dy < 0 {
			g.memScroll++
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
			g.memScroll -= MemViewRows
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
			g.memScroll += MemViewRows
		}
		if g.memScroll < 0 {
			g.memScroll = 0
		}
		if g.memScroll > g.memScrollMax {
			g.memScroll = g.memScrollMax
		}
	}

	if g.paused {
		return nil
	}

	// Key polling (under CPU lock)
	g.CPU.Mu.Lock()
	for ebKey, chipKey := range keyMap {
		g.CPU.Keys[chipKey] = ebiten.IsKeyPressed(ebKey)
	}
	// NOTE: CPU steps are driven by the CPU goroutine in main.go.
	// Update() only handles input + timer decrement sync with ebiten's 60Hz.
	g.CPU.Mu.Unlock()

	// Frame jitter tracking
	now := time.Now()
	if !g.lastFrame.IsZero() {
		elapsed := now.Sub(g.lastFrame).Seconds() * 1000
		g.frameJitter = elapsed - (1000.0 / 60.0)
		if g.frameJitter < 0 {
			g.frameJitter = -g.frameJitter
		}
	}
	g.lastFrame = now

	return nil
}

// ── Draw ──────────────────────────────────────────────────────────────────────
func (g *Gui) Draw(screen *ebiten.Image) {
	screen.Fill(colBg)

	g.CPU.Mu.Lock()
	snap := g.CPU.Snapshot()
	displayCopy := g.CPU.Display
	g.CPU.Mu.Unlock()

	// ── Render CHIP-8 display ──
	img := image.NewRGBA(image.Rect(0, 0, ChipW, ChipH))
	for i, pixel := range displayCopy {
		x := i % ChipW
		y := i / ChipW
		if pixel == 1 {
			img.SetRGBA(x, y, colPixelOn)
		} else {
			img.SetRGBA(x, y, colPixelOff)
		}
	}
	g.offscreen.WritePixels(img.Pix)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(Upscale), float64(Upscale))
	screen.DrawImage(g.offscreen, op)

	if !g.DebugOn {
		// Minimal HUD
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("IPS:%-4d  PC:0x%03X  [ESC=pause]", g.CPU.GetIPS(), snap.PC),
			4, GameH-16)
		return
	}

	// ── Debug Panel ───────────────────────────────────────────────────────────
	g.drawDebugPanel(screen, snap)
}

func (g *Gui) drawDebugPanel(screen *ebiten.Image, snap cpu.Snapshot) {
	px := GameW + 10
	lh := 14 // line height
	y := 8

	drawLabel := func(label, val string, vc color.RGBA) {
		ebitenutil.DebugPrintAt(screen, label, px, y)
		ebitenutil.DebugPrintAt(screen, val, px+80, y)
		y += lh
	}
	drawLine := func(text string, c color.RGBA) {
		_ = c // DebugPrint is always white; color param reserved for future font renderer
		ebitenutil.DebugPrintAt(screen, text, px, y)
		y += lh
	}
	drawSep := func() {
		ebitenutil.DebugPrintAt(screen, "------------------------------------", px, y)
		y += lh
	}

	// ── Header ──
	drawLine("  CHIP-8 DEBUG PANEL", colHdr)
	drawLine(fmt.Sprintf("  ROM: %s", truncate(g.RomPath, 28)), colDim)
	if g.paused {
		drawLine("  [ PAUSED — ESC to resume ]", colHighlight)
	}
	drawSep()

	// ── PC / SP / I ──
	drawLine("REGISTERS", colHdr)
	drawLabel("PC:", fmt.Sprintf("0x%03X", snap.PC), colHighlight)
	drawLabel("SP:", fmt.Sprintf("%d", snap.SP), colValue)
	drawLabel("I: ", fmt.Sprintf("0x%03X", snap.I), colValue)
	drawLabel("DT:", fmt.Sprintf("%d", snap.DelayTimer), colValue)
	drawLabel("ST:", fmt.Sprintf("%d", snap.SoundTimer), colValue)
	y += 4
	drawSep()

	// ── V0–VF ──
	drawLine("V REGISTERS", colHdr)
	for i := 0; i < 16; i++ {
		label := fmt.Sprintf("V%X:", i)
		val := fmt.Sprintf("0x%02X  (%3d)", snap.V[i], snap.V[i])
		ebitenutil.DebugPrintAt(screen, label, px, y)
		ebitenutil.DebugPrintAt(screen, val, px+36, y)
		y += lh
	}
	y += 4
	drawSep()

	// ── Current opcode ──
	drawLine("CURRENT OPCODE", colHdr)
	opcodeStr := cpu.Disassemble(snap.PC, snap.LastOpcode)
	drawLine(fmt.Sprintf("  %04X  %s", snap.LastOpcode, opcodeStr), colValue)
	y += 4
	drawSep()

	// ── Metrics ──
	drawLine("METRICS", colHdr)
	drawLabel("IPS:", fmt.Sprintf("%d", g.CPU.GetIPS()), colIPS)
	drawLabel("Jitter:", fmt.Sprintf("%.2fms", g.frameJitter), colValue)
	y += 4
	drawSep()

	// ── Memory Viewer ──
	drawLine("MEMORY VIEWER  [scroll=wheel PgUp/Dn]", colHdr)

	g.CPU.Mu.Lock()
	startByte := g.memScroll * MemBytesRow
	mem := g.CPU.MemorySlice(startByte, MemViewRows*MemBytesRow)
	pcAddr := int(snap.PC)
	g.CPU.Mu.Unlock()

	for row := 0; row < MemViewRows && row*MemBytesRow < len(mem); row++ {
		addr := startByte + row*MemBytesRow
		line := fmt.Sprintf("%04X: ", addr)
		for col := 0; col < MemBytesRow && row*MemBytesRow+col < len(mem); col++ {
			b := mem[row*MemBytesRow+col]
			byteAddr := addr + col
			if byteAddr == pcAddr || byteAddr == pcAddr+1 {
				line += fmt.Sprintf("[%02X]", b)
			} else {
				line += fmt.Sprintf(" %02X ", b)
			}
		}
		ebitenutil.DebugPrintAt(screen, line, px, y)
		y += lh
	}

	// scroll indicator
	total := 4096 / MemBytesRow
	pct := 0
	if total > 0 {
		pct = g.memScroll * 100 / total
	}
	drawLine(fmt.Sprintf("  addr 0x%04X  (%d%%)", startByte, pct), colDim)

	// ── Key state ──
	y += 4
	drawSep()
	drawLine("KEYS", colHdr)
	g.CPU.Mu.Lock()
	var keyStr string
	for i := 0; i < 16; i++ {
		if g.CPU.Keys[i] {
			keyStr += fmt.Sprintf("%X ", i)
		}
	}
	g.CPU.Mu.Unlock()
	if keyStr == "" {
		keyStr = "none"
	}
	drawLine("  "+keyStr, colValue)
}

// ── Layout ────────────────────────────────────────────────────────────────────
func (g *Gui) Layout(outsideWidth, outsideHeight int) (int, int) {
	if g.DebugOn {
		return TotalW, GameH
	}
	return GameW, GameH
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return "..." + s[len(s)-max+3:]
}