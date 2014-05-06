package main

import (
	"github.com/adkennan/Go-SDL/gfx"
	"github.com/adkennan/Go-SDL/sdl"
	"github.com/adkennan/Go-SDL/ttf"
	"image/color"
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

type Window interface {
	CreateSurface(w, h int) Surface
	DrawSurface(x, y int, sfc Surface)
	DrawSurfacePart(dx, dy int, sfc Surface, sx, sy, sw, sh int)
	Clear(c color.Color)
	ClearRect(c color.Color, x, y, w, h int)
	Update()
	W() int
	H() int
	SetClipRect(x, y, w, h int)
	ClearClipRect()
}

type Surface interface {
	Clear()
	SetColor(c color.Color)
	DrawSurfacePart(dx, dy int, sfc Surface, sx, sy, sw, sh int)
	DrawSurface(x, y int, sfc Surface)
	DrawLine(x1, y1, x2, y2 int)
	DrawPoint(x, y int)
	Fill(x1, y1, x2, y2 int)
	Update()
	W() int
	H() int
}

type sdlWindow struct {
	b    *MessageBroker
	win  *sdl.Surface
	w, h int
}

func (this *sdlWindow) CreateSurface(w, h int) Surface {

	s := sdl.CreateRGBSurface(sdl.SWSURFACE,
		int(w),
		int(h),
		32,
		0x000000ff,
		0x0000ff00,
		0x00ff0000,
		0xff000000)

	return &sdlSurface{s, w, h, color.RGBA{0, 0, 0, 255}}
}

func (this *sdlWindow) SetClipRect(x, y, w, h int) {
	this.win.SetClipRect(&sdl.Rect{int16(x), int16(y), uint16(w), uint16(h)})
}

func (this *sdlWindow) ClearClipRect() {
	this.win.SetClipRect(nil)
}

func (this *sdlWindow) DrawSurface(x, y int, sfc Surface) {

	ss := sfc.(*sdlSurface)
	dst := &sdl.Rect{int16(x), int16(y), uint16(ss.w), uint16(ss.h)}

	this.win.Blit(dst, ss.s, nil)
}

func (this *sdlWindow) DrawSurfacePart(dx, dy int, sfc Surface, sx, sy, sw, sh int) {

	ss := sfc.(*sdlSurface)
	src := &sdl.Rect{int16(sx), int16(sy), uint16(sw), uint16(sh)}
	dst := &sdl.Rect{int16(dx), int16(dy), uint16(sw), uint16(sh)}

	this.win.Blit(dst, ss.s, src)
}

func toSdlColor(format *sdl.PixelFormat, c color.Color) uint32 {
	r, g, b, a := c.RGBA()

	return sdl.MapRGBA(format, uint8(r), uint8(g), uint8(b), uint8(a))
}

func (this *sdlWindow) Clear(c color.Color) {

	this.win.FillRect(nil, toSdlColor(this.win.Format, c))
}

func (this *sdlWindow) ClearRect(c color.Color, x, y, w, h int) {

	this.win.FillRect(&sdl.Rect{int16(x), int16(y), uint16(w), uint16(h)}, toSdlColor(this.win.Format, c))
}

func (this *sdlWindow) Close() {
}

func (this *sdlWindow) Update() {

	this.win.Flip()
}

func (this *sdlWindow) W() int {
	return this.w
}

func (this *sdlWindow) H() int {
	return this.h
}

func (this *sdlWindow) runEventLoop() {

	sdl.EnableUNICODE(1)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					var r rune
					if e.Keysym.Unicode != 0 {
						rs := utf16.Decode([]uint16{e.Keysym.Unicode})
						r = rs[0]
					}
					m := &KeyMessage{MessageBase{MT_KeyPress}, e.Keysym.Sym, r}
					this.b.Publish(m)
				}
			}
		}
		sdl.Delay(20)
	}

	this.b.PublishId(MT_Quit)
	sdl.Quit()
}

func newWindow(broker *MessageBroker) Window {

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		panic(sdl.GetError())
	}

	i := sdl.GetVideoInfo()
	w := int(i.Current_w)
	h := int(i.Current_h)

	s := sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE)

	sdl.WM_ToggleFullScreen(s)

	sdl.ShowCursor(0)

	win := &sdlWindow{broker, s, w, h}

	go win.runEventLoop()

	return win
}

type sdlSurface struct {
	s    *sdl.Surface
	w, h int
	c    color.Color
}

func (this *sdlSurface) W() int {
	return this.w
}

func (this *sdlSurface) H() int {
	return this.h
}

func (this *sdlSurface) DrawSurface(x, y int, sfc Surface) {
	ss := sfc.(*sdlSurface)
	dst := &sdl.Rect{int16(x), int16(y), uint16(ss.w), uint16(ss.h)}

	this.s.Blit(dst, ss.s, nil)
}

func (this *sdlSurface) DrawSurfacePart(dx, dy int, sfc Surface, sx, sy, sw, sh int) {

	ss := sfc.(*sdlSurface)
	src := &sdl.Rect{int16(sx), int16(sy), uint16(sw), uint16(sh)}
	dst := &sdl.Rect{int16(dx), int16(dy), uint16(sw), uint16(sh)}

	this.s.Blit(dst, ss.s, src)
}

func (this *sdlSurface) Free() {
}

func (this *sdlSurface) Clear() {
	this.s.FillRect(nil, 0)
}

func (this *sdlSurface) SetColor(c color.Color) {
	this.c = c
}

func (this *sdlSurface) DrawLine(x1, y1, x2, y2 int) {
	gfx.LineColor(this.s, int16(x1), int16(y1), int16(x2), int16(y2), toSdlColor(this.s.Format, this.c))
}

func (this *sdlSurface) DrawPoint(x, y int) {
	gfx.PixelColor(this.s, int16(x), int16(y), toSdlColor(this.s.Format, this.c))
}

func (this *sdlSurface) Update() {
	this.s.UpdateRect(0, 0, uint32(this.w), uint32(this.h))
}

func (this *sdlSurface) Fill(x1, y1, x2, y2 int) {
	r, g, b, a := this.c.RGBA()
	gfx.BoxRGBA(this.s, int16(x1), int16(y1), int16(x2), int16(y2),
		uint8(r), uint8(g), uint8(b), uint8(a))
}
