package main

import (
	//"bytes"
	//"github.com/jackyb/go-sdl2/sdl"
	//"image"
	"math"
)

const (
	screenModeText = iota
	screenModeSplit
	screenModeGraphic
)

const splitScreenSize = 4

type Screen struct {
	screen      Window
	drawSurface Surface
	w, h        int64
	isDirty     bool
	ws          *Workspace
	screenMode  int
}

func initScreen(workspace *Workspace) *Screen {

	ss := newWindow()
	w := ss.W()
	h := ss.H()
	ds := ss.CreateSurface(w, h)

	s := &Screen{ss, ds, int64(w), int64(h), true, workspace, screenModeSplit}

	workspace.registerBuiltIn("TEXT", "", 3, _s_Text)
	workspace.registerBuiltIn("FULLSCREEN", "FS", 0, _s_Fullscreen)
	workspace.registerBuiltIn("TEXTSCREEN", "TS", 0, _s_Textscreen)
	workspace.registerBuiltIn("SPLITSCREEN", "SS", 0, _s_Splitscreen)

	return s
}

func (this *Screen) Update() {
	if !this.isDirty {
		return
	}

	t := this.ws.turtle
	this.screen.Clear(t.screenColor)

	gm := this.ws.glyphMap
	gs := this.drawSurface
	c := this.ws.console
	cs := c.Surface()

	switch this.screenMode {
	case screenModeGraphic:
		this.screen.DrawSurface(0, 0, gs)
		if t.turtleState == turtleStateShown {
			this.DrawTurtle()
		}
	case screenModeText:
		this.screen.DrawSurface(0, 0, cs)
	case screenModeSplit:
		th := gm.charHeight * splitScreenSize

		this.screen.DrawSurfacePart(0, 0, gs, 0, 0, int(this.w), int(this.h)-th)
		this.screen.DrawSurfacePart(0, int(this.h)-th, cs,
			0, (1+c.FirstLineOfSplitScreen())*gm.charHeight, int(this.w), th)

		if t.turtleState == turtleStateShown {
			this.DrawTurtle()
		}
	}

	this.screen.Update()
	this.isDirty = false
}

func (this *Screen) Close() {
	//this.screen.Close()
}

func (this *Screen) clear() {
	this.drawSurface.SetColor(this.ws.turtle.screenColor)
	this.drawSurface.Clear()
	this.isDirty = true
}

func (this *Screen) DrawTurtle() {
	t := this.ws.turtle
	x := int(t.x+float64(this.w/2)) - turtleSize
	y := int(-t.y+float64(this.h/2)) - turtleSize

	this.screen.DrawSurface(x, y, t.sprite)

	/*
		t := this.ws.turtle
		d := normAngle(t.d)
		x := t.x + float64(this.w/2)
		y := -t.y + float64(this.h/2)
		x1 := int(x - (5 * math.Cos((d-90)*dToR)))
		y1 := int(y - (5 * math.Sin((d-90)*dToR)))
		x2 := int(x - (5 * math.Cos((d+90)*dToR)))
		y2 := int(y - (5 * math.Sin((d+90)*dToR)))
		x3 := int(x - (10 * math.Cos(d*dToR)))
		y3 := int(y - (10 * math.Sin(d*dToR)))

		r := this.screen
		r.SetColor(colorWhite)
		r.DrawLine(x1, y1, x2, y2)
		r.DrawLine(x2, y2, x3, y3)
		r.DrawLine(x3, y3, x1, y1)
	*/
}

func (this *Screen) normX(x int64) int64 {
	return x + this.w/2
}

func (this *Screen) normY(y int64) int64 {
	return -y + this.h/2
}

func (this *Screen) DrawLine(x1, y1, x2, y2 int64) {

	x1 = this.normX(x1)
	x2 = this.normX(x2)
	y1 = this.normY(y1)
	y2 = this.normY(y2)

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

	r := this.drawSurface
	r.SetColor(this.ws.turtle.penColor)
	for {
		//print("x=", x1, ", y=", y1, "\n")
		r.DrawPoint(int(x1), int(y1))
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

func _s_Text(frame Frame, parameters []Node) (Node, error) {
	/*
		fx, fy, err := evalNumericParams(parameters[0], parameters[1])
		if err != nil {
			return nil, err
		}
		buf := &bytes.Buffer{}
		nodeToText(buf, parameters[2], false)
		t := buf.String()
		ws := frame.workspace()
		nx := ws.screen.normX(int64(fx))
		ny := ws.screen.normY(int64(fy))

		for _, c := range t {
			nx = int64(ws.glyphMap.renderGlyph(c, glyphStyleNormal, ws.screen.drawSurface, int(nx), int(ny)))
		}

		ws.screen.StateChanged()
	*/return nil, nil
}

func _s_Fullscreen(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()
	ws.screen.screenMode = screenModeGraphic
	ws.screen.StateChanged()

	return nil, nil
}

func _s_Textscreen(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()
	ws.screen.screenMode = screenModeText
	ws.screen.StateChanged()

	return nil, nil
}

func _s_Splitscreen(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()
	ws.screen.screenMode = screenModeSplit
	ws.screen.StateChanged()

	return nil, nil
}
