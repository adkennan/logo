package main

import (
	"image/color"
	"math"
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
}

func (this *Turtle) onDrawLine(x1, y1, x2, y2 int64) {
	this.ws.screen.DrawLine(x1, y1, x2, y2)
}

func (this *Turtle) onStateChanged() {
	this.ws.screen.StateChanged()
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
}

func initTurtle(ws *Workspace) *Turtle {
	turtle := &Turtle{
		0, 0, 0, 1.0, turtleStateShown, penStateDown, penModePaint, colorWhite, colorBlack, colorWhite, ws, nil}

	turtle.sprite = ws.screen.screen.CreateSurface(turtleSize*2, turtleSize*2)
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

	if t.penState == penStateDown {
		t.onDrawLine(int64(t.x), int64(t.y), int64(x2), int64(y2))
	}

	t.x = x2
	t.y = y2

	t.onStateChanged()

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

	if t.penState == penStateDown {
		t.onDrawLine(int64(t.x), int64(t.y), int64(x2), int64(y2))
	}

	t.x = x2
	t.y = y2

	t.onStateChanged()

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

	t.onStateChanged()
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

	t.onStateChanged()
	return nil, nil
}

func _t_ShowTurtle(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	t.turtleState = turtleStateShown

	t.onStateChanged()
	return nil, nil
}

func _t_HideTurtle(frame Frame, parameters []Node) (Node, error) {
	t := frame.workspace().turtle
	t.turtleState = turtleStateHidden

	t.onStateChanged()
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

	frame.workspace().screen.clear()

	frame.workspace().turtle.onStateChanged()
	return nil, nil
}

func _t_Home(frame Frame, parameters []Node) (Node, error) {

	t := frame.workspace().turtle
	if t.penState == penStateDown {
		t.onDrawLine(int64(t.x), int64(t.y), 0, 0)
	}

	t.x = 0
	t.y = 0
	t.d = 0

	t.updateSprite()

	t.onStateChanged()
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

	t.onStateChanged()
	return nil, nil

}

func _t_SetX(frame Frame, parameters []Node) (Node, error) {

	x, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	if t.penState == penStateDown {
		t.onDrawLine(int64(t.x), int64(t.y), int64(x), int64(t.y))
	}

	t.x = x

	t.onStateChanged()
	return nil, nil
}

func _t_SetY(frame Frame, parameters []Node) (Node, error) {

	y, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	t := frame.workspace().turtle
	if t.penState == penStateDown {
		t.onDrawLine(int64(t.x), int64(t.y), int64(t.x), int64(y))
	}

	t.y = y

	t.onStateChanged()
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
		if t.penState == penStateDown {
			t.onDrawLine(int64(t.x), int64(t.y), int64(x), int64(y))
		}

		t.x = x
		t.y = y

		t.onStateChanged()
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
