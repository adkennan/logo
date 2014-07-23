package main

import (
	"github.com/adkennan/Go-SDL/gfx"
	"github.com/adkennan/Go-SDL/sdl"
	"github.com/adkennan/Go-SDL/ttf"
	"image/color"
	"path"
	"unicode/utf16"
	"unsafe"
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
		g = &sdlSurface{gs, uintptr(unsafe.Pointer(gs.Pixels)), int(gs.W), int(gs.H), color.RGBA{0, 0, 0, 0}, 0}
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
	CreateSurface(w, h int, withAlpha bool) Surface
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

func (this *sdlWindow) CreateSurface(w, h int, withAlpha bool) Surface {

	var a uint32 = 0x000000ff
	if withAlpha {
		a = 0x000000ff
	}

	s := sdl.CreateRGBSurface(sdl.SWSURFACE,
		int(w),
		int(h),
		32,
		0xff000000,
		0x00ff0000,
		0x0000ff00,
		a)

	c := color.RGBA{0, 0, 0, 255}
	return &sdlSurface{s, uintptr(unsafe.Pointer(s.Pixels)), w, h, c, toSdlColor(s.Format, c)}
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
	s      *sdl.Surface
	pixels uintptr
	w, h   int
	c      color.Color
	sdlCol uint32
}

func (this *sdlSurface) setPixel(x, y int, c uint32) {

	*((*uint32)(unsafe.Pointer(this.pixels +
		uintptr((y*this.w+x)*4)))) = c
}

func (this *sdlSurface) getPixel(x, y int) uint32 {

	return *((*uint32)(unsafe.Pointer(this.pixels +
		uintptr((y*this.w+x)*4))))
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
	this.sdlCol = toSdlColor(this.s.Format, c)
}

func (this *sdlSurface) DrawLine(x1, y1, x2, y2 int) {
	gfx.LineColor(this.s, int16(x1), int16(y1), int16(x2), int16(y2), this.sdlCol)
}

func (this *sdlSurface) DrawPoint(x, y int) {
	if x < 0 || x >= this.w || y < 0 || y >= this.h {
		return
	}
	this.setPixel(x, y, this.sdlCol)
}

func (this *sdlSurface) ErasePoint(x, y int) {
	if x < 0 || x >= this.w || y < 0 || y >= this.h {
		return
	}
	this.setPixel(x, y, 0)
}

func (this *sdlSurface) ReversePoint(x, y int) {
	if x < 0 || x >= this.w || y < 0 || y >= this.h {
		return
	}
	r1, g1, b1, _ := this.s.At(x, y).RGBA()
	r2, g2, b2, a := this.c.RGBA()

	c := color.RGBA{uint8(r1 ^ r2), uint8(g1 ^ g2), uint8(b1 ^ b2), uint8(a)}

	this.setPixel(x, y, toSdlColor(this.s.Format, c))
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
	sdlCol := this.getPixel(x, y)
	var r, g, b, a uint8
	sdl.GetRGBA(sdlCol, this.s.Format, &r, &g, &b, &a)
	return color.RGBA{r, g, b, a}
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

	tc := this.getPixel(x, y)
	if tc == this.sdlCol {
		return x, y, x, y
	}

	sweep := uintptr(this.w * 4)

	q := &fillNode{nil, x, y}
	for q != nil {
		sx, sy := q.x, q.y
		p := this.pixels + uintptr((sy*this.w+sx)*4)
		q = q.n

		x1 := sx

		lc := *(*uint32)(unsafe.Pointer(p))
		for x1 > 0 && lc == tc {
			x1--
			p -= uintptr(4)
			lc = *(*uint32)(unsafe.Pointer(p))
		}

		x2 := x1 + 1
		p += uintptr(4)
		ux := -1
		dx := -1
		lc = *(*uint32)(unsafe.Pointer(p))
		for x2 < this.w-1 && lc == tc {
			if sy > 0 {
				uc := *(*uint32)(unsafe.Pointer(p - sweep))
				if uc == tc {
					if ux != x2-1 {
						q = &fillNode{q, x2, sy - 1}
					}
					ux = x2
				}
			}
			if sy < this.h-1 {
				dc := *(*uint32)(unsafe.Pointer(p + sweep))
				if dc == tc {
					if dx != x2-1 {
						q = &fillNode{q, x2, sy + 1}
					}
					dx = x2
				}
			}
			*(*uint32)(unsafe.Pointer(p)) = this.sdlCol
			p += uintptr(4)
			x2++
			lc = *(*uint32)(unsafe.Pointer(p))
		}

		if x1+1 < x2-1 {
			if x1 < minX {
				minX = x1
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
	}

	return
}
