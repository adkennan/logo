package main

import (
	"github.com/adkennan/Go-SDL/gfx"
	"github.com/adkennan/Go-SDL/sdl"
	"github.com/adkennan/Go-SDL/ttf"
	"image/color"
	"path"
	"unicode/utf16"
)

var resourceDir string = "res"
var normalFontName string = "DejaVuSansMono.ttf"
var boldFontName string = "DejaVuSansMono-Bold.ttf"

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
		ttf.OpenFont(path.Join(resourceDir, normalFontName), fontSize),
		ttf.OpenFont(path.Join(resourceDir, boldFontName), fontSize)}

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
	ErasePoint(x, y int)
	ReversePoint(x, y int)
	ColorAt(x, y int) color.Color
	Fill(x1, y1, x2, y2 int)
	FillTriangle(x1, y1, x2, y2, x3, y3 int)
	Flood(x, y int) (minX, minY, maxX, maxY int)
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

	s := sdl.CreateRGBSurface(sdl.HWSURFACE,
		int(w),
		int(h),
		32,
		0xff000000,
		0x00ff0000,
		0x0000ff00,
		0x000000ff)

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

	//	gfx.RectangleRGBA(this.win, int16(dx), int16(dy), int16(dx+sw), int16(dy+sh), 0, 255, 0, 255)
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

const (
	keyRepeatFirst = 20
	keyRepeatSubs  = 3
)

func (this *sdlWindow) runEventLoop() {

	sdl.EnableUNICODE(1)
	defer func() {
		this.b.PublishId(MT_Quit)
		sdl.Quit()
	}()

	running := true
	var km *KeyMessage = nil
	keyCount := 0
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				switch e.Type {
				case sdl.KEYDOWN:
					var r rune
					if e.Keysym.Unicode != 0 {
						rs := utf16.Decode([]uint16{e.Keysym.Unicode})
						r = rs[0]
					}
					km = &KeyMessage{MessageBase{MT_KeyPress}, e.Keysym.Sym, e.Keysym.Mod, r}
					this.b.Publish(km)
					keyCount = keyRepeatFirst

				case sdl.KEYUP:
					km = nil
				}
			}
		}
		if keyCount == 0 {
			if km != nil {
				this.b.Publish(km)
				keyCount = keyRepeatSubs
			}
		} else {
			keyCount--
		}
		sdl.Delay(20)
	}

}

func newWindow(broker *MessageBroker, ww, wh int) Window {

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		panic(sdl.GetError())
	}

	i := sdl.GetVideoInfo()
	w := int(i.Current_w)
	h := int(i.Current_h)

	if ww == 0 || wh == 0 {
		ww = w
		wh = h
	}
	s := sdl.SetVideoMode(ww, wh, 32, sdl.RESIZABLE)

	if ww == w && wh == h {
		sdl.WM_ToggleFullScreen(s)
	}

	sdl.ShowCursor(0)

	win := &sdlWindow{broker, s, ww, wh}

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

func (this *sdlSurface) ErasePoint(x, y int) {
	gfx.PixelColor(this.s, int16(x), int16(y), 0x000000FF)
}

func (this *sdlSurface) ReversePoint(x, y int) {
	if x < 0 || x >= this.w || y < 0 || y >= this.h {
		return
	}
	r1, g1, b1, _ := this.s.At(x, y).RGBA()
	r2, g2, b2, a := this.c.RGBA()

	c := color.RGBA{uint8(r1 ^ r2), uint8(g1 ^ g2), uint8(b1 ^ b2), uint8(a)}

	gfx.PixelColor(this.s, int16(x), int16(y), toSdlColor(this.s.Format, c))
}

func (this *sdlSurface) Update() {
	this.s.UpdateRect(0, 0, uint32(this.w), uint32(this.h))
}

func (this *sdlSurface) Fill(x1, y1, x2, y2 int) {
	r, g, b, a := this.c.RGBA()
	gfx.BoxRGBA(this.s, int16(x1), int16(y1), int16(x2), int16(y2),
		uint8(r), uint8(g), uint8(b), uint8(a))
}

func (this *sdlSurface) FillTriangle(x1, y1, x2, y2, x3, y3 int) {
	gfx.FilledTrigonColor(this.s, int16(x1), int16(y1), int16(x2), int16(y2), int16(x3), int16(y3), toSdlColor(this.s.Format, this.c))
}

func (this *sdlSurface) ColorAt(x, y int) color.Color {
	if x < 0 || x >= this.w || y < 0 || y >= this.h {
		return color.RGBA{}
	}
	return this.s.At(x, y)
}

type fillNode struct {
	n    *fillNode
	x, y int
}

func colorEqual(c1, c2 color.Color) bool {

	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2
}

func (this *sdlSurface) Flood(x, y int) (minX, minY, maxX, maxY int) {

	minX, minY, maxX, maxY = this.w, this.h, 0, 0

	sdlC := toSdlColor(this.s.Format, this.c)

	tc := this.s.At(x, y)
	if colorEqual(tc, this.c) {
		return x, y, x, y
	}
	q := &fillNode{nil, x, y}

	for q != nil {
		sx, sy := q.x, q.y
		q = q.n
		x2 := sx
		ux := -1
		dx := -1

		lc := this.s.At(x2, sy)
		for x2 >= 0 && colorEqual(lc, tc) {
			if sy > 0 {
				uc := this.s.At(x2, sy-1)
				if colorEqual(uc, tc) {
					if ux != x2-1 {
						q = &fillNode{q, x2, sy - 1}
					}
					ux = x2
				}
			}
			if sy < this.h-1 {
				dc := this.s.At(x2, sy+1)
				if colorEqual(dc, tc) {
					if dx != x2-1 {
						q = &fillNode{q, x2, sy + 1}
					}
					dx = x2
				}
			}
			x2--
			lc = this.s.At(x2, sy)
		}
		fx := x2
		x2 = sx + 1
		ux = -1
		dx = -1
		lc = this.s.At(x2, sy)
		for x2 < this.w-1 && colorEqual(lc, tc) {
			if sy > 0 {
				uc := this.s.At(x2, sy-1)
				if colorEqual(uc, tc) {
					if ux != x2+1 {
						q = &fillNode{q, x2, sy - 1}
					}
					ux = x2
				}
			}
			if sy < this.h-1 {
				dc := this.s.At(x2, sy+1)
				if colorEqual(dc, tc) {
					if dx != x2+1 {
						q = &fillNode{q, x2, sy + 1}
					}
					dx = x2
				}
			}
			x2++
			lc = this.s.At(x2, sy)
		}
		gfx.LineColor(this.s, int16(fx+1), int16(sy), int16(x2-1), int16(sy), sdlC)
		if fx < minX {
			minX = fx
		}
		if x2 > maxX {
			maxX = x2
		}
		if sy < minY {
			minY = sy
		}
		if sy > maxY {
			maxY = sy
		}
	}

	return
}
