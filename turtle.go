package main

import (
	"bytes"
	"fmt"
	"image/color"
	"math"
	"strings"
	"sync"
	"time"
)

const (
	penStateUp int = iota
	penStateDown
	penStateReverse
	penStateErase
	turtleStateShown
	turtleStateHidden
	borderModeWindow
	borderModeFence
	borderModeWrap
)

var penStateNames [4]string = [4]string{"PENUP", "PENDOWN", "PENREVERSE", "PENERASE"}

const turtleSize = 14

var (
	colorBlack      = color.RGBA{0, 0, 0, 0xff}
	colorMagenta    = color.RGBA{0x8c, 0x3b, 0x50, 0xff}
	colorDarkBlue   = color.RGBA{0x55, 0x45, 0x84, 0xff}
	colorPurple     = color.RGBA{0xfc, 0x56, 0xea, 0xff}
	colorDarkGreen  = color.RGBA{0x00, 0x67, 0x53, 0xff}
	colorGrey       = color.RGBA{0x90, 0x90, 0x90, 0xff}
	colorMediumBlue = color.RGBA{0x00, 0xa3, 0xeb, 0xff}
	colorLightBlue  = color.RGBA{0xcc, 0xbf, 0xf4, 0xff}
	colorBrown      = color.RGBA{0x4c, 0x5c, 0x20, 0xff}
	colorOrange     = color.RGBA{0xf6, 0x7e, 0x34, 0xff}
	colorPink       = color.RGBA{0xff, 0xb6, 0xc7, 0xff}
	colorGreen      = color.RGBA{0x00, 0xc9, 0x43, 0xff}
	colorYellow     = color.RGBA{0xc5, 0xd2, 0x9d, 0xff}
	colorAqua       = color.RGBA{0x86, 0xdb, 0xc9, 0xff}
	colorWhite      = color.RGBA{0xff, 0xff, 0xff, 0xff}
	turtleColor     = color.RGBA{0x21, 0xd0, 0x21, 0xff}

	colorsMap = map[string]color.RGBA{
		"black":      colorBlack,
		"magenta":    colorMagenta,
		"darkblue":   colorDarkBlue,
		"purple":     colorPurple,
		"darkgreen":  colorDarkGreen,
		"grey":       colorGrey,
		"mediumblue": colorMediumBlue,
		"lightblue":  colorLightBlue,
		"brown":      colorBrown,
		"orange":     colorOrange,
		"pink":       colorPink,
		"green":      colorGreen,
		"yellow":     colorYellow,
		"aqua":       colorAqua,
		"white":      colorWhite,
	}
)

const dToR float64 = math.Pi / 180.0

type Turtle struct {
	x, y         float64
	d            float64
	renderedD    int
	scale        float64
	turtleState  int
	penState     int
	borderMode   int
	penColor     color.RGBA
	screenColor  color.RGBA
	floodColor   color.RGBA
	ws           *Workspace
	sprite       Surface
	image        Surface
	channel      *Channel
	dirtyRegions []*Region
	mutex        *sync.Mutex
	visW         int
	visH         int
}

func initTurtle(ws *Workspace) *Turtle {
	turtle := &Turtle{
		0, 0, 0, -1, 1.0, turtleStateShown, penStateDown, borderModeWindow,
		colorWhite, colorBlack, colorWhite, ws, nil, nil, nil, nil, &sync.Mutex{}, 0, 0}

	turtle.sprite = ws.screen.screen.CreateSurface(turtleSize*2, turtleSize*2, true)
	turtle.image = ws.screen.screen.CreateSurface(ws.screen.screen.W(), ws.screen.screen.H(), true)
	turtle.channel = ws.broker.Subscribe("Turtle", MT_VisibleAreaChange)
	turtle.dirtyRegions = make([]*Region, 0, 16)

	ws.registerBuiltIn("FORWARD", "FD", 1, _t_Forward)
	ws.registerBuiltIn("BACK", "BK", 1, _t_Back)
	ws.registerBuiltIn("RIGHT", "RT", 1, _t_Right)
	ws.registerBuiltIn("LEFT", "LT", 1, _t_Left)

	ws.registerBuiltIn("CLEARSCREEN", "CS", 0, _t_ClearScreen)
	ws.registerBuiltIn("HOME", "", 0, _t_Home)
	ws.registerBuiltIn("SETPOS", "", 1, _t_SetPos)
	ws.registerBuiltIn("SETHEADING", "SETH", 1, _t_SetHeading)
	ws.registerBuiltIn("SETX", "", 1, _t_SetX)
	ws.registerBuiltIn("SETY", "", 1, _t_SetY)
	ws.registerBuiltIn("SHOWTURTLE", "ST", 0, _t_ShowTurtle)
	ws.registerBuiltIn("HIDETURTLE", "HT", 0, _t_HideTurtle)
	ws.registerBuiltIn("PENUP", "PU", 0, _t_PenUp)
	ws.registerBuiltIn("PENDOWN", "PD", 0, _t_PenDown)
	ws.registerBuiltIn("PENERASE", "PE", 0, _t_PenErase)
	ws.registerBuiltIn("PENREVERSE", "PX", 0, _t_PenReverse)
	ws.registerBuiltIn("SETPC", "", 1, _t_SetPc)
	ws.registerBuiltIn("SETBG", "", 1, _t_SetBg)
	ws.registerBuiltIn("PENCOLOR", "PC", 0, _t_PenColor)
	ws.registerBuiltIn("BACKGROUND", "BG", 0, _t_Background)
	ws.registerBuiltIn("PEN", "", 0, _t_Pen)
	ws.registerBuiltIn("FILL", "", 0, _t_Fill)

	ws.registerBuiltIn("HEADING", "", 0, _t_Heading)
	ws.registerBuiltIn("POS", "", 0, _t_Pos)
	ws.registerBuiltIn("SHOWNP", "", 0, _t_Shownp)
	ws.registerBuiltIn("TOWARDS", "", 1, _t_Towards)
	ws.registerBuiltIn("XCOR", "", 0, _t_XCor)
	ws.registerBuiltIn("YCOR", "", 0, _t_YCor)
	ws.registerBuiltIn("TEXT", "", 3, _t_Text)

	ws.registerBuiltIn("CLEAN", "", 0, _t_Clean)
	ws.registerBuiltIn("DOT", "", 1, _t_Dot)
	ws.registerBuiltIn("DOTP", "", 1, _t_Dotp)

	ws.registerBuiltIn("WINDOW", "", 0, _t_Window)
	ws.registerBuiltIn("WRAP", "", 0, _t_Wrap)
	ws.registerBuiltIn("FENCE", "", 0, _t_Fence)

	go turtle.listen()
	go turtle.tick()

	return turtle
}

func (this *Turtle) normX(x int) int {
	return x + int(this.image.W())/2
}

func (this *Turtle) denormX(x int) int {
	return x - int(this.image.W())/2
}

func (this *Turtle) normY(y int) int {
	return -y + int(this.image.H())/2
}

func (this *Turtle) denormY(y int) int {
	return this.normY(y)
}

func (this *Turtle) clear() {
	this.image.SetColor(this.screenColor)
	this.image.Clear()

	this.invalidate()
}

func (this *Turtle) invalidate() {
	this.addDirtyRegion(0, 0, this.visW, this.visH)
}

func (this *Turtle) addDirtyRegion(x1, y1, x2, y2 int) {

	rx1 := intMin(x1, x2)
	ry1 := intMin(y1, y2)
	rx2 := intMax(x1, x2)
	ry2 := intMax(y1, y2)

	r := &Region{rx1, ry1, rx2 - rx1, ry2 - ry1}

	this.mutex.Lock()
	defer this.mutex.Unlock()

	for _, o := range this.dirtyRegions {
		if o.Overlaps(r) {
			o.Combine(r)
			return
		}
	}
	this.dirtyRegions = append(this.dirtyRegions, r)
}

func (this *Turtle) tick() {
	for {
		this.mutex.Lock()

		if len(this.dirtyRegions) > 0 {

			regions := make([]*Region, 0, len(this.dirtyRegions))
			for _, r := range this.dirtyRegions {
				regions = append(regions, r.Clone())
			}
			this.dirtyRegions = this.dirtyRegions[:0]
			this.channel.Publish(newRegionMessage(MT_UpdateGfx, this.image, regions))
		}

		this.mutex.Unlock()
		time.Sleep(30 * time.Millisecond)
	}
}

func (this *Turtle) listen() {
	for m := this.channel.Wait(); m != nil; m = this.channel.Wait() {
		switch rm := m.(type) {
		case *VisibleAreaChangeMessage:
			this.visW = rm.w
			this.visH = rm.h
		}
	}
}

func (this *Turtle) offScreen() bool {
	x := this.normX(int(this.x))
	y := this.normY(int(this.y))

	return x < 0 || x >= this.visW || y < 0 || y >= this.visH
}

func (this *Turtle) drawLine(x1, y1, x2, y2 int) (int, int) {
	x1 = this.normX(x1)
	x2 = this.normX(x2)
	y1 = this.normY(y1)
	y2 = this.normY(y2)

	rx1 := x1
	ry1 := y1

	dx := int(math.Abs(float64(x2 - x1)))
	dy := int(math.Abs(float64(y2 - y1)))

	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	r := this.image
	w := this.visW
	h := this.visH

	r.SetColor(this.penColor)
	for {
		switch this.penState {
		case penStateDown:
			r.DrawPoint(x1, y1)
		case penStateErase:
			r.ErasePoint(x1, y1)
		case penStateReverse:
			r.ReversePoint(x1, y1)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx

			if x1 < 0 || x1 >= w {
				switch this.borderMode {
				case borderModeFence:
					goto done
				case borderModeWrap:

					this.addDirtyRegion(rx1, ry1, x1, y1)

					tx := x1 - x2

					if x1 < 0 {
						x1 = w - 1
					} else {
						x1 = 0
					}
					x2 = x1 - tx
					rx1 = x1
					ry1 = y1
				default:
					break
				}
			}
		}

		if e2 < dx {
			err += dx
			y1 += sy

			if y1 < 0 || y1 >= h {

				switch this.borderMode {
				case borderModeFence:
					goto done
				case borderModeWrap:

					this.addDirtyRegion(rx1, ry1, x1, y1)

					ty := y1 - y2

					if y1 < 0 {
						y1 = h - 1
					} else {
						y1 = 0
					}
					y2 = y1 - ty
					rx1 = x1
					ry1 = y1
				default:
					break
				}
			}
		}
	}
done:

	if this.penState != penStateUp {
		this.addDirtyRegion(rx1, ry1, x2, y2)
	} else {
		this.addDirtyRegion(rx1, ry1, rx1, ry1)
		this.addDirtyRegion(x2, y2, x2, y2)
	}

	return this.denormX(x1), this.denormY(y1)
}

func (this *Turtle) fill() {

	x := this.normX(int(this.x))
	y := this.normY(int(this.y))

	this.image.SetColor(this.penColor)
	x1, y1, x2, y2 := this.image.Flood(x, y)
	this.addDirtyRegion(x1, y1, x2, y2+1)
}

func (this *Turtle) updateSprite() {

	t := this
	d := normAngle(t.d)
	ht := float64(turtleSize / 2)
	x1 := int(turtleSize - (ht * math.Cos((d-90)*dToR)))
	y1 := int(turtleSize - (ht * math.Sin((d-90)*dToR)))
	x2 := int(turtleSize - (ht * math.Cos((d+90)*dToR)))
	y2 := int(turtleSize - (ht * math.Sin((d+90)*dToR)))
	x3 := int(turtleSize - (float64(turtleSize) * math.Cos(d*dToR)))
	y3 := int(turtleSize - (float64(turtleSize) * math.Sin(d*dToR)))

	tx := this.normX(int(this.x)) - turtleSize
	ty := this.normY(int(this.y)) - turtleSize
	r := this.sprite
	r.Clear()
	r.SetColor(turtleColor)
	r.FillTriangle(x1, y1, x2, y2, x3, y3)

	this.addDirtyRegion(tx-turtleSize*2, ty-turtleSize*2, tx+turtleSize*2, ty+turtleSize*2)
	this.renderedD = int(d)
}

func (this *Turtle) refreshTurtle() {

	tx := this.normX(int(this.x))
	ty := this.normY(int(this.y))
	this.addDirtyRegion(tx-turtleSize, ty-turtleSize, tx+turtleSize, ty+turtleSize)
}

func (this *Turtle) spriteNeedsUpdate() bool {

	return int(this.d) != this.renderedD
}

func normAngle(d float64) float64 {
	d = 90 - d
	for d < 0 {
		d += 360
	}
	for d > 359 {
		d -= 360
	}
	return d
}

func _t_Forward(frame Frame, parameters []Node) *CallResult {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	x2 := t.x - math.Cos(normAngle(t.d)*dToR)*delta
	y2 := t.y + math.Sin(normAngle(t.d)*dToR)*delta

	t.refreshTurtle()
	x, y := t.drawLine(int(t.x), int(t.y), int(x2), int(y2))

	if x != int(x2) {
		x2 = float64(x)
	}
	if y != int(y2) {
		y2 = float64(y)
	}

	t.x = x2
	t.y = y2
	t.refreshTurtle()

	return nil
}

func _t_Back(frame Frame, parameters []Node) *CallResult {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	x2 := t.x + math.Cos(normAngle(t.d)*dToR)*delta
	y2 := t.y - math.Sin(normAngle(t.d)*dToR)*delta

	t.refreshTurtle()

	x, y := t.drawLine(int(t.x), int(t.y), int(x2), int(y2))

	if x != int(x2) {
		x2 = float64(x)
	}
	if y != int(y2) {
		y2 = float64(y)
	}

	t.x = x2
	t.y = y2
	t.refreshTurtle()

	return nil
}

func _t_Left(frame Frame, parameters []Node) *CallResult {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	t.d += delta
	for t.d < 0 {
		t.d += 360
	}

	return nil
}

func _t_Right(frame Frame, parameters []Node) *CallResult {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	t.d -= delta
	for t.d >= 360 {
		t.d -= 360
	}

	return nil
}

func _t_ShowTurtle(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.turtleState = turtleStateShown

	t.refreshTurtle()
	return nil
}

func _t_HideTurtle(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.turtleState = turtleStateHidden

	t.refreshTurtle()
	return nil
}

func _t_PenUp(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.penState = penStateUp

	return nil
}

func _t_PenDown(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.penState = penStateDown

	return nil
}

func _t_ClearScreen(frame Frame, parameters []Node) *CallResult {

	_t_Home(frame, parameters)

	frame.workspace().turtle.clear()

	return nil
}

func _t_Home(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	t.refreshTurtle()
	t.drawLine(int(t.x), int(t.y), 0, 0)

	t.x = 0
	t.y = 0
	t.d = 0

	return nil
}

func _t_SetHeading(frame Frame, parameters []Node) *CallResult {

	d, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	t.d = d
	for t.d >= 360 {
		t.d -= 360
	}

	return nil

}

func _t_SetX(frame Frame, parameters []Node) *CallResult {

	x, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	t.refreshTurtle()
	x2, _ := t.drawLine(int(t.x), int(t.y), int(x), int(t.y))

	if x2 != int(x) {
		x = float64(x2)
	}

	t.x = x
	t.refreshTurtle()

	return nil
}

func _t_SetY(frame Frame, parameters []Node) *CallResult {

	y, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	t.refreshTurtle()
	_, y2 := t.drawLine(int(t.x), int(t.y), int(t.x), int(y))

	if y2 != int(y) {
		y = float64(y2)
	}

	t.y = y
	t.refreshTurtle()

	return nil
}

func parseCoords(node Node) (x, y float64, err error) {
	switch l := node.(type) {
	case *ListNode:

		if l.length() != 2 {
			return 0, 0, err
		}

		x, err := evalToNumber(l.firstChild)
		if err != nil {
			return 0, 0, err
		}
		y, err := evalToNumber(l.firstChild.next())
		if err != nil {
			return 0, 0, err
		}

		return x, y, nil
	}

	return 0, 0, errorListExpected(node)
}

func _t_SetPos(frame Frame, parameters []Node) *CallResult {

	x, y, err := parseCoords(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	t := frame.workspace().turtle
	t.refreshTurtle()

	x2, y2 := t.drawLine(int(t.x), int(t.y), int(x), int(y))

	if x2 != int(x) {
		x = float64(x2)
	}

	if y2 != int(y) {
		y = float64(y2)
	}

	t.x = x
	t.y = y

	t.refreshTurtle()
	return nil
}

func _t_Heading(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	return returnResult(createNumericNode(t.d))
}

func _t_Pos(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	fn := createNumericNode(t.x)
	fn.addNode(createNumericNode(t.y))

	return returnResult(newListNode(-1, -1, fn))
}

func _t_Shownp(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	if t.turtleState == turtleStateShown {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _t_Towards(frame Frame, parameters []Node) *CallResult {
	return nil
}

func _t_XCor(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	return returnResult(createNumericNode(t.x))
}

func _t_YCor(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	return returnResult(createNumericNode(t.y))
}

func _t_Text(frame Frame, parameters []Node) *CallResult {

	fx, fy, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}
	buf := &bytes.Buffer{}
	nodeToText(buf, parameters[2], false)
	text := buf.String()

	gm := frame.workspace().glyphMap
	t := frame.workspace().turtle
	nx := t.normX(int(fx))
	ny := t.normY(int(fy))

	x1 := nx

	for _, c := range text {
		nx = gm.renderGlyph(c, glyphStyleNormal, t.image, nx, ny)
	}

	t.addDirtyRegion(x1, ny, nx, ny+gm.charHeight)

	return nil
}

func evalToColorPart(n Node) (uint8, error) {
	v, err := evalToNumber(n)
	if err != nil {
		return 0, err
	}

	if v < 0 || v > 255 {
		return 0, errorNumberNotInRange(n, 0, 255)
	}

	return uint8(v), nil
}

func _t_SetPc(frame Frame, parameters []Node) *CallResult {

	c, err := evalToColor(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	frame.workspace().turtle.penColor = c
	return nil
}

func _t_SetBg(frame Frame, parameters []Node) *CallResult {

	c, err := evalToColor(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	frame.workspace().turtle.screenColor = c
	frame.workspace().turtle.invalidate()

	return nil
}

func _t_PenColor(frame Frame, parameters []Node) *CallResult {

	return returnResult(colorToNode(frame.workspace().turtle.penColor))
}

func _t_Background(frame Frame, parameters []Node) *CallResult {
	return returnResult(colorToNode(frame.workspace().turtle.screenColor))
}

func evalToColor(node Node) (color.RGBA, error) {

	switch p := node.(type) {
	case *WordNode:
		cc, ok := colorsMap[strings.ToLower(p.value)]
		if !ok {
			return cc, errorUnknownColor(p, p.value)
		}
		return cc, nil

	case *ListNode:
		if p.length() != 3 {
			return color.RGBA{}, errorListOfNItemsExpected(p, 3)
		}
		n := p.firstChild
		r, err := evalToColorPart(n)
		if err != nil {
			return color.RGBA{}, err
		}
		n = n.next()
		g, err := evalToColorPart(n)
		if err != nil {
			return color.RGBA{}, err
		}
		n = n.next()
		b, err := evalToColorPart(n)
		if err != nil {
			return color.RGBA{}, err
		}

		return color.RGBA{r, g, b, 0xff}, nil
	}

	return color.RGBA{}, nil
}

func colorToNode(c color.Color) Node {
	r, g, b, _ := c.RGBA()

	rn := newWordNode(-1, -1, fmt.Sprint(r>>8), true)
	gn := newWordNode(-1, -1, fmt.Sprint(g>>8), true)
	bn := newWordNode(-1, -1, fmt.Sprint(b>>8), true)

	rn.addNode(gn)
	gn.addNode(bn)

	return newListNode(-1, -1, rn)
}

func _t_Clean(frame Frame, parameters []Node) *CallResult {

	frame.workspace().turtle.clear()

	return nil
}

func _t_Dot(frame Frame, parameters []Node) *CallResult {

	switch l := parameters[0].(type) {
	case *ListNode:

		ll, err := evalList(frame, l)
		if err != nil {
			return errorResult(err)
		}
		if ll.length() != 2 {
			return errorResult(errorListOfNItemsExpected(l, 2))
		}

		x, err := evalToNumber(ll.firstChild)
		if err != nil {
			return errorResult(err)
		}
		y, err := evalToNumber(ll.firstChild.next())
		if err != nil {
			return errorResult(err)
		}

		t := frame.workspace().turtle

		t.drawLine(int(x), int(y), int(x), int(y))

		return nil
	}

	return errorResult(errorListExpected(parameters[0]))
}

func _t_Fence(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.borderMode = borderModeFence

	if t.offScreen() {
		t.x = 0
		t.y = 0
		t.refreshTurtle()
	}
	return nil
}

func _t_Wrap(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.borderMode = borderModeWrap

	if t.offScreen() {
		t.x = 0
		t.y = 0
		t.refreshTurtle()
	}
	return nil
}

func _t_Window(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.borderMode = borderModeWindow

	return nil
}

func _t_Fill(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	t.fill()
	return nil
}

func _t_PenErase(frame Frame, parameters []Node) *CallResult {

	t := frame.workspace().turtle
	t.penState = penStateErase
	return nil
}

func _t_PenReverse(frame Frame, parameters []Node) *CallResult {
	t := frame.workspace().turtle
	t.penState = penStateReverse
	return nil
}

func _t_Pen(frame Frame, parameters []Node) *CallResult {
	return returnResult(newWordNode(-1, -1, penStateNames[frame.workspace().turtle.penState], true))
}

func _t_Dotp(frame Frame, parameters []Node) *CallResult {

	x, y, err := parseCoords(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	t := frame.workspace().turtle

	xx := t.normX(int(x))
	yy := t.normY(int(y))
	r, g, b, _ := t.image.ColorAt(xx, yy).RGBA()
	if r == 0 && g == 0 && b == 0 {
		return returnResult(falseNode)
	}
	return returnResult(trueNode)
}
