package main

import (
	"github.com/adkennan/Go-SDL/gfx"
	"github.com/adkennan/Go-SDL/sdl"
	"image/color"
)

type Window interface {
	CreateSurface(w, h int) Surface
	DrawSurface(x, y int, sfc Surface)
	DrawSurfacePart(dx, dy int, sfc Surface, sx, sy, sw, sh int)
	Clear(c color.Color)
	Update()
	W() int
	H() int
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

func newWindow() Window {

	i := sdl.GetVideoInfo()
	w := int(i.Current_w)
	h := int(i.Current_h)

	s := sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE)

	sdl.WM_ToggleFullScreen(s)

	sdl.ShowCursor(0)

	win := &sdlWindow{s, w, h}

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
