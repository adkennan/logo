package logo

import (
	"github.com/adkennan/Go-SDL/gfx"
	"github.com/adkennan/Go-SDL/sdl"
	"image"
	"image/color"
	"math"
)

type Screen struct {
	screen      *sdl.Surface
	drawSurface *sdl.Surface
	drawImage   *image.RGBA
	w, h        int64
	isDirty     bool
	ws          *Workspace
}

func initScreen(workspace *Workspace, w, h int) *Screen {
	ss := sdl.SetVideoMode(w, h, 32, sdl.RESIZABLE)

	di := image.NewRGBA(image.Rect(0, 0, w-1, h-1))
	ds := sdl.CreateRGBSurfaceFrom(di.Pix, w, h, 32, di.Stride, 0x000000ff, 0x0000ff00, 0x00ff0000, 0xff000000)

	s := &Screen{ss, ds, di, int64(w), int64(h), true, workspace}

	return s
}

func toSdlColor(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return (r << 24) & (g << 16) & (b << 8) & 0xff
}

func (this *Screen) Update() {
	if !this.isDirty {
		return
	}

	t := this.ws.turtle

	this.screen.FillRect(nil, toSdlColor(t.screenColor))
	this.screen.Blit(nil, this.drawSurface, nil)
	if t.turtleState == turtleStateShown {
		this.DrawTurtle()
	}
	this.screen.Flip()

	this.isDirty = false
}

func (this *Screen) clear() {
	this.drawSurface.FillRect(nil, 0x00000000)
	this.isDirty = true
}

func (this *Screen) DrawTurtle() {

	t := this.ws.turtle
	d := normAngle(t.d)
	x := t.x + float64(this.w/2)
	y := -t.y + float64(this.h/2)
	x1 := int16(x - (5 * math.Cos((d-90)*dToR)))
	y1 := int16(y - (5 * math.Sin((d-90)*dToR)))
	x2 := int16(x - (5 * math.Cos((d+90)*dToR)))
	y2 := int16(y - (5 * math.Sin((d+90)*dToR)))
	x3 := int16(x - (10 * math.Cos(d*dToR)))
	y3 := int16(y - (10 * math.Sin(d*dToR)))

	gfx.LineColor(this.screen, x1, y1, x2, y2, 0xFFFFFFFF)
	gfx.LineColor(this.screen, x2, y2, x3, y3, 0xFFFFFFFF)
	gfx.LineColor(this.screen, x3, y3, x1, y1, 0xFFFFFFFF)
}

func (this *Screen) DrawLine(x1, y1, x2, y2 int64) {

	this.drawSurface.Lock()
	defer func() {
		this.drawSurface.Unlock()
		this.drawSurface.UpdateRect(0, 0, 0, 0)
	}()

	x1 += this.w / 2
	x2 += this.w / 2
	y1 = -y1 + this.h/2
	y2 = -y2 + this.h/2

	dx := int64(math.Abs(float64(x2 - x1)))
	dy := int64(math.Abs(float64(y2 - y1)))

	sx := int64(-1)
	if x1 < x2 {
		sx = 1
	}
	sy := int64(-1)
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		//print("x=", x1, ", y=", y1, "\n")
		this.drawImage.Set(int(x1), int(y1), this.ws.turtle.penColor)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
			if (sx == -1 && x1 == 0) || (sx == 1 && x1 == this.w-1) {
				break
			}
		}
		if e2 < dx {
			err += dx
			y1 += sy
			if (sy == -1 && y1 == 0) || (sy == 1 && y1 == this.h-1) {
				break
			}
		}
	}

	this.isDirty = true
}

func (this *Screen) StateChanged() {
	this.isDirty = true
}
