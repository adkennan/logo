package main

import (
	//	"fmt"
	//"github.com/adkennan/Go-SDL/gfx"
	"github.com/adkennan/Go-SDL/sdl"
	"github.com/adkennan/Go-SDL/ttf"
	"image/color"
	"unicode"
	"unicode/utf16"
)

var normalFontName string = "res/DejaVuSansMono.ttf"
var boldFontName string = "res/DejaVuSansMono-Bold.ttf"

const (
	glyphStyleNormal  = 0x01
	glyphStyleBold    = 0x02
	glyphStyleInverse = 0x04

	fontSize = 24
)

type GlyphMap struct {
	fonts      []*ttf.Font
	glyphs     map[int]map[rune]*sdlSurface
	charHeight int
	charWidth  int
}

var textColFg sdl.Color = sdl.Color{0xFF, 0xFF, 0xFF, 0xFF}
var textColBg sdl.Color = sdl.Color{0x00, 0x00, 0x00, 0xFF}

func initGlyphMap() *GlyphMap {

	ttf.Init()

	fs := []*ttf.Font{
		ttf.OpenFont(normalFontName, fontSize),
		ttf.OpenFont(boldFontName, fontSize)}

	gm := &GlyphMap{fs, make(map[int]map[rune]*sdlSurface), fs[0].Height(), 0}

	g := gm.getGlyph('e', glyphStyleNormal)
	gm.charWidth = int(g.W())

	return gm
}

func (this *GlyphMap) getGlyph(c rune, glyphStyle int) Surface {

	gm, exists := this.glyphs[glyphStyle]
	if !exists {
		gm = make(map[rune]*sdlSurface, 30)
		this.glyphs[glyphStyle] = gm
	}

	g, exists := gm[c]
	if !exists {
		fgc := textColFg
		bgc := textColBg
		if (glyphStyle & glyphStyleInverse) == glyphStyleInverse {
			fgc = textColBg
			bgc = textColFg
		}

		s := string([]rune{c})
		var gs *sdl.Surface
		if (glyphStyle & glyphStyleNormal) == glyphStyleNormal {
			gs = ttf.RenderUTF8_Shaded(this.fonts[0], s, fgc, bgc)
		} else {
			gs = ttf.RenderUTF8_Shaded(this.fonts[1], s, fgc, bgc)
		}
		g = &sdlSurface{gs, int(gs.W), int(gs.H), color.RGBA{0, 0, 0, 0}}
		gm[c] = g
	}

	return g
}

func (this *GlyphMap) renderGlyph(c rune, glyphStyle int, dst Surface, x, y int) int {
	g := this.getGlyph(c, glyphStyle)

	dst.DrawSurface(x, y, g)

	return x + int(g.W())
}

type ConsoleScreen struct {
	ws    *Workspace
	sfcs  [2]Surface
	sfcIx int
	cx    int
	cy    int
	in    chan *sdl.Keysym
}

func initConsole(workspace *Workspace, w, h int) *ConsoleScreen {
	sfcs := [2]Surface{
		workspace.screen.screen.CreateSurface(w, h),
		workspace.screen.screen.CreateSurface(w, h),
	}
	cs := &ConsoleScreen{
		workspace,
		sfcs,
		0,
		0,
		0,
		make(chan *sdl.Keysym, 100)}

	cs.sfcs[0].Clear()
	cs.sfcs[1].Clear()

	cs.sfcs[0].SetColor(color.RGBA{0, 0, 0, 255})
	cs.sfcs[0].Fill(0, 0, w, h)

	cs.sfcs[1].SetColor(color.RGBA{0, 0, 0, 255})
	cs.sfcs[1].Fill(0, 0, w, h)

	workspace.registerBuiltIn("CLEARTEXT", "CT", 0, _c_ClearText)

	return cs
}

func (this *ConsoleScreen) Id() int {
	return 0
}

func (this *ConsoleScreen) Name() string {
	return ""
}

func (this *ConsoleScreen) Input() chan *sdl.Keysym {
	return this.in
}

func (this *ConsoleScreen) FirstLineOfSplitScreen() int {

	if this.cy <= splitScreenSize {
		return 0
	}
	return this.cy - splitScreenSize
}

func (this *ConsoleScreen) Clear() {

	s := this.sfcs[this.sfcIx]
	s.SetColor(color.RGBA{0, 0, 0, 255})
	s.Fill(0, 0, s.W(), s.H())

	this.cx = 0
	this.cy = 0

	this.ws.screen.StateChanged()
}

func (this *ConsoleScreen) ReadLine() (string, error) {
	cursorPos := 0
	chars := make([]rune, 0, 10)
	this.drawEditLine(cursorPos, chars)

	for ks := range this.in {
		switch ks.Sym {
		case sdl.K_RETURN:
			line := string(chars)

			this.clearEditLine()
			this.Write(line)
			this.Write("\n")
			return line, nil
		case sdl.K_LEFT:
			if cursorPos > 0 {
				cursorPos--
			}
		case sdl.K_RIGHT:
			if cursorPos < len(chars) {
				cursorPos++
			}
		case sdl.K_HOME:
			cursorPos = 0
		case sdl.K_END:
			cursorPos = len(chars)
		case sdl.K_BACKSPACE:
			if cursorPos > 0 {
				chars = append(chars[:cursorPos-1], chars[cursorPos:]...)
				cursorPos--
			}
		case sdl.K_DELETE:
			if cursorPos < len(chars)-1 {
				chars = append(chars[:cursorPos], chars[cursorPos+1:]...)
			}
		default:
			if ks.Unicode != 0 {
				rs := utf16.Decode([]uint16{ks.Unicode})
				r := rs[0]
				if unicode.IsGraphic(r) {
					if cursorPos == len(chars) {
						chars = append(chars, r)
					} else {
						chars = append(chars[:cursorPos],
							append([]rune{r}, chars[cursorPos:]...)...)
					}
					cursorPos++
				}
			}
		}
		this.drawEditLine(cursorPos, chars)
	}
	return "", nil
}

func (this *ConsoleScreen) clearEditLine() {
	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	dst := this.sfcs[this.sfcIx]

	dst.SetColor(color.RGBA{0, 0, 0, 255})
	dst.Fill(nx, ny, dst.W(), ny+gm.charHeight)
}

func (this *ConsoleScreen) drawEditLine(cursorPos int, chars []rune) {

	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	dst := this.sfcs[this.sfcIx]

	this.clearEditLine()

	for _, c := range chars {
		nx = gm.renderGlyph(c, glyphStyleNormal, dst, nx, ny)
	}

	dst.SetColor(color.RGBA{255, 255, 255, 255})
	dst.Fill(this.cx+((cursorPos+2)*gm.charWidth)+2,
		ny+gm.charHeight-2,
		this.cx+((cursorPos+3)*gm.charWidth)-2,
		ny+gm.charHeight)

	this.ws.screen.StateChanged()
}

func (this *ConsoleScreen) Surface() Surface {
	return this.sfcs[this.sfcIx]
}

func (this *ConsoleScreen) Write(text string) error {

	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	for _, c := range text {
		if unicode.IsGraphic(c) {
			nx = gm.renderGlyph(c, glyphStyleNormal, this.sfcs[this.sfcIx], nx, ny)
			this.cx++

		} else if c == '\n' {
			this.cx = 0
			nx = 0
			this.cy++
			ny += gm.charHeight
			if ny+gm.charHeight >= this.sfcs[0].H() {

				src := this.sfcs[this.sfcIx]
				dst := this.sfcs[1-this.sfcIx]

				dst.SetColor(color.RGBA{0, 0, 0, 255})
				dst.Fill(0, 0, dst.W(), dst.H())

				dst.DrawSurface(0, -gm.charHeight, src)

				this.sfcIx = 1 - this.sfcIx

				this.cy--
				ny -= gm.charHeight
			}
		}
	}

	this.ws.screen.StateChanged()

	return nil
}

func (this *ConsoleScreen) Close() error {
	return nil
}

func (this *ConsoleScreen) IsInteractive() bool {
	return true
}

func _c_ClearText(frame Frame, parameters []Node) (Node, error) {

	frame.workspace().console.Clear()

	return nil, nil
}
