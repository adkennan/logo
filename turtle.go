package main

import (
	"bytes"
	"image/color"
	"math"
	"sync"
	"time"
)

const (
	penStateUp int = iota
	penStateDown
	penModePaint
	penModeReverse
	penModeErase
	turtleStateShown
	turtleStateHidden
)

const turtleSize = 16
const gridSize = 32

var (
	colorBlack = color.RGBA{0, 0, 0, 0xff}
	colorWhite = color.RGBA{0xff, 0xff, 0xff, 0xff}
)

const dToR float64 = math.Pi / 180.0

type Turtle struct {
	x, y        float64
	d           float64
	scale       float64
	turtleState int
	penState    int
	penMode     int
	penColor    color.RGBA
	screenColor color.RGBA
	floodColor  color.RGBA
	ws          *Workspace
	sprite      Surface
	image       Surface
	broker      *MessageBroker
	dirtyGrid   []bool
	gridSweep   int
	mutex       *sync.Mutex
}

func (this *Turtle) normX(x int) int {
	return x + int(this.image.W())/2
}

func (this *Turtle) normY(y int) int {
	return -y + int(this.image.H())/2
}

func (this *Turtle) clear() {
	this.image.SetColor(this.screenColor)
	this.image.Clear()

	this.addDirtyRegion(0, 0, this.image.W(), this.image.H())
}

func (this *Turtle) addDirtyRegion(x1, y1, x2, y2 int) {

	gridHeight := len(this.dirtyGrid) / this.gridSweep
	bx1 := intMax(0, intMin(x1, x2)/gridSize)
	by1 := intMax(0, intMin(y1, y2)/gridSize)
	bx2 := intMin(this.gridSweep-1, (intMax(x1, x2) / gridSize))
	by2 := intMin(gridHeight-1, (intMax(y1, y2) / gridSize))

	this.mutex.Lock()
	defer this.mutex.Unlock()

	for y := by1; y <= by2; y++ {
		l := this.dirtyGrid[y*this.gridSweep : (y+1)*this.gridSweep]
		for x := bx1; x <= bx2; x++ {
			//if x >= bx1 && x <= bx2 {
			l[x] = true
			//}
		}
	}
}

func (this *Turtle) tick() {
	for {
		this.mutex.Lock()

		gridHeight := this.image.H() / gridSize
		regions := make([]*Region, 0, 1)
		for y := 0; y < gridHeight; y++ {
			l := this.dirtyGrid[y*this.gridSweep : (y+1)*this.gridSweep]
			for x := 0; x < this.gridSweep; x++ {
				if l[x] {
					l[x] = false
					for _, r := range regions {
						if r.ContainsPoint(x, y) {
							goto matched
						} else if r.AdjacentTo(x, y) {
							r.ExpandToInclude(x, y)
							goto matched
						}
					}
					regions = append(regions, &Region{x, y, 1, 1})
				} else {
				}
			matched:
			}
		}

		if len(regions) > 0 {
			for _, r := range regions {
				r.Multiply(gridSize)
			}
			this.broker.Publish(newRegionMessage(MT_UpdateGfx, this.image, regions))
		}

		this.mutex.Unlock()
		time.Sleep(30 * time.Millisecond)
	}
}

func (this *Turtle) drawLine(x1, y1, x2, y2 int) {
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
	w := r.W()
	h := r.H()
	r.SetColor(this.penColor)
	for {
		if this.penState == penStateDown {
			r.DrawPoint(x1, y1)
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
			if (sx == -1 && x1 == 0) || (sx == 1 && x1 == w-1) {
				break
			}
		}
		if e2 < dx {
			err += dx
			y1 += sy
			if (sy == -1 && y1 == 0) || (sy == 1 && y1 == h-1) {
				break
			}
		}
	}

	if this.penState == penStateDown {
		this.addDirtyRegion(rx1, ry1, x2, y2)
	} else {
		this.addDirtyRegion(rx1, ry1, rx1, ry1)
		this.addDirtyRegion(x2, y2, x2, y2)
	}
}

func (this *Turtle) updateSprite() {

	t := this
	d := normAngle(t.d)
	x1 := int(turtleSize - (5 * math.Cos((d-90)*dToR)))
	y1 := int(turtleSize - (5 * math.Sin((d-90)*dToR)))
	x2 := int(turtleSize - (5 * math.Cos((d+90)*dToR)))
	y2 := int(turtleSize - (5 * math.Sin((d+90)*dToR)))
	x3 := int(turtleSize - (10 * math.Cos(d*dToR)))
	y3 := int(turtleSize - (10 * math.Sin(d*dToR)))

	r := this.sprite
	r.Clear()
	r.SetColor(colorWhite)
	r.DrawLine(x1, y1, x2, y2)
	r.DrawLine(x2, y2, x3, y3)
	r.DrawLine(x3, y3, x1, y1)

	this.addDirtyRegion(this.normX(int(this.x))-turtleSize, this.normY(int(this.y))-turtleSize,
		this.normX(int(this.x))+turtleSize, this.normY(int(this.y))+turtleSize)
}

func (this *Turtle) refreshTurtle() {

	tx := this.normX(int(this.x))
	ty := this.normY(int(this.y))
	this.addDirtyRegion(tx-turtleSize/2, ty-turtleSize/2, tx+turtleSize/2, ty+turtleSize/2)
}

func initTurtle(ws *Workspace) *Turtle {
	turtle := &Turtle{
		0, 0, 0, 1.0, turtleStateShown, penStateDown, penModePaint, colorWhite, colorBlack, colorWhite, ws, nil, nil, nil, nil, 0, &sync.Mutex{}}

	turtle.sprite = ws.screen.screen.CreateSurface(turtleSize*2, turtleSize*2)
	turtle.image = ws.screen.screen.CreateSurface(ws.screen.screen.W(), ws.screen.screen.H())
	turtle.broker = ws.broker
	turtle.gridSweep = turtle.image.W() / 32
	turtle.dirtyGrid = make([]bool, turtle.gridSweep*(turtle.image.H()/32))
	turtle.updateSprite()

	ws.registerBuiltIn("FORWARD", "FD", 1, _t_Forward)
	ws.registerBuiltIn("BACK", "BK", 1, _t_Back)
	ws.registerBuiltIn("RIGHT", "RT", 1, _t_Right)
	ws.registerBuiltIn("LEFT", "LT", 1, _t_Left)

	ws.registerBuiltIn("CLEARSCREEN", "CS", 0, _t_ClearScreen)
	ws.registerBuiltIn("HOME", "", 0, _t_Home)
	ws.registerBuiltIn("SETPOS", "", 1, _t_SetPos)
	ws.registerBuiltIn("SETHEADING", "SETH", 2, _t_SetHeading)
	ws.registerBuiltIn("SETX", "", 1, _t_SetX)
	ws.registerBuiltIn("SETY", "", 1, _t_SetY)
	ws.registerBuiltIn("SHOWTURTLE", "ST", 0, _t_ShowTurtle)
	ws.registerBuiltIn("HIDETURTLE", "HT", 0, _t_HideTurtle)
	ws.registerBuiltIn("PENUP", "PU", 0, _t_PenUp)
	ws.registerBuiltIn("PENDOWN", "PD", 0, _t_PenDown)

	ws.registerBuiltIn("HEADING", "", 0, _t_Heading)
	ws.registerBuiltIn("POS", "", 0, _t_Pos)
	ws.registerBuiltIn("SHOWNP", "", 0, _t_Shownp)
	ws.registerBuiltIn("TOWARDS", "", 1, _t_Towards)
	ws.registerBuiltIn("XCOR", "", 0, _t_XCor)
	ws.registerBuiltIn("YCOR", "", 0, _t_YCor)
	ws.registerBuiltIn("TEXT", "", 3, _t_Text)

	go turtle.tick()

	return turtle
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

func _t_Forward(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	x2 := t.x - math.Cos(normAngle(t.d)*dToR)*delta
	y2 := t.y + math.Sin(normAngle(t.d)*dToR)*delta

	t.drawLine(int(t.x), int(t.y), int(x2), int(y2))

	t.x = x2
	t.y = y2

	return nil, nil
}

func _t_Back(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	x2 := t.x + math.Cos(normAngle(t.d)*dToR)*delta
	y2 := t.y - math.Sin(normAngle(t.d)*dToR)*delta

	t.drawLine(int(t.x), int(t.y), int(x2), int(y2))

	t.x = x2
	t.y = y2

	return nil, nil
}

func _t_Left(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	t.d += delta
	for t.d < 0 {
		t.d += 360
	}

	t.updateSprite()

	return nil, nil
}

func _t_Right(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	t.d -= delta
	for t.d >= 360 {
		t.d -= 360
	}

	t.updateSprite()

	return nil, nil
}

func _t_ShowTurtle(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	t.turtleState = turtleStateShown

	t.refreshTurtle()
	return nil, nil
}

func _t_HideTurtle(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	t.turtleState = turtleStateHidden

	t.refreshTurtle()
	return nil, nil
}

func _t_PenUp(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	t.penState = penStateUp

	return nil, nil
}

func _t_PenDown(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	t.penState = penStateDown

	return nil, nil
}

func _t_ClearScreen(frame Frame, parameters []Node) (Node, error) {

	_t_Home(frame, parameters)

	frame.workspace().turtle.clear()

	return nil, nil
}

func _t_Home(frame Frame, parameters []Node) (Node, error) {

	t := frame.workspace().turtle
	t.drawLine(int(t.x), int(t.y), 0, 0)

	t.x = 0
	t.y = 0
	t.d = 0

	t.updateSprite()

	return nil, nil
}

func _t_SetHeading(frame Frame, parameters []Node) (Node, error) {

	d, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	t.d = d
	for t.d >= 360 {
		t.d -= 360
	}

	t.updateSprite()

	return nil, nil

}

func _t_SetX(frame Frame, parameters []Node) (Node, error) {

	x, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	t.drawLine(int(t.x), int(t.y), int(x), int(t.y))

	t.x = x

	return nil, nil
}

func _t_SetY(frame Frame, parameters []Node) (Node, error) {

	y, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	t.drawLine(int(t.x), int(t.y), int(t.x), int(y))

	t.y = y

	return nil, nil
}

func _t_SetPos(frame Frame, parameters []Node) (Node, error) {

	switch l := parameters[0].(type) {
	case *ListNode:

		if l.length() != 2 {
			return nil, errorListOfNItemsExpected(l, 2)
		}

		x, err := evalToNumber(l.firstChild)
		if err != nil {
			return nil, err
		}
		y, err := evalToNumber(l.firstChild.next())
		if err != nil {
			return nil, err
		}

		t := frame.workspace().turtle
		t.drawLine(int(t.x), int(t.y), int(x), int(y))

		t.x = x
		t.y = y

		return nil, nil
	}

	return nil, errorListExpected(parameters[0])
}

func _t_Heading(frame Frame, parameters []Node) (Node, error) {

	t := frame.workspace().turtle
	return createNumericNode(t.d), nil
}

func _t_Pos(frame Frame, parameters []Node) (Node, error) {

	t := frame.workspace().turtle
	fn := createNumericNode(t.x)
	fn.addNode(createNumericNode(t.y))

	return newListNode(-1, -1, fn), nil
}

func _t_Shownp(frame Frame, parameters []Node) (Node, error) {

	t := frame.workspace().turtle
	if t.turtleState == turtleStateShown {
		return trueNode, nil
	}
	return falseNode, nil
}

func _t_Towards(frame Frame, parameters []Node) (Node, error) {
	return nil, nil
}

func _t_XCor(frame Frame, parameters []Node) (Node, error) {

	t := frame.workspace().turtle
	return createNumericNode(t.x), nil
}

func _t_YCor(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	return createNumericNode(t.y), nil
}

func _t_Text(frame Frame, parameters []Node) (Node, error) {

	fx, fy, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
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

	return nil, nil
}
