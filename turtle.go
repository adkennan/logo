package logo

const (
	penStateUp int = iota
	penStateDown
	penModePaint
	penModeReverse
	penModeErase
)

const (
	colorBlack = 0x000000FF
	colorWhite = 0xFFFFFFFF
	colorGrey  = 0x888888FF
	colorRed   = 0xFF0000FF
	colorGreen = 0x00FF00FF
	colorBlue  = 0x0000FFFF
)

const dToR float64 = math.Pi / 180.0

type Turtle struct {
	x, y        int64
	d           float64
	scale       float64
	penState    int
	penMode     int
	penColor    uint32
	screenColor uint32
	floodColor  uint32
	penWidth    int
	penHeight   int
}

type TurtleEvents struct {
	DrawLine     func(x1, y1, x2, y2 int64)
	StateChanged func(turtle *Turtle)
}

var turtle *Turtle
var turtleEvents *TurtleEvents

func onDrawLine(x1, y1, x2, y2 int64) {
	if turtleEvents != nil {
		turtleEvents.DrawLine(x1, y1, x2, y2)
	}
}

func onStateChanged() {
	if turtleEvents != nil {
		turtleEvents.StateChanged(turtle)
	}
}

func InitTurtle() {
	turtle = &Turtle{
		0, 0, 0.0, 1.0, penStateDown, penModePaint, colorWhite, colorBlack, colorWhite, 1, 1}
}

func _t_Forward(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	x2 := int64(math.Cos(delta*turtle.d*dToR)) + turtle.x
	y2 := int64(math.Sin(delta*turtle.d*dToR)) + turtle.y

	if turtle.penState == penDown {
		onDrawLine(turtle.x, turtle.y, x2, y2)
	}

	turtle.x = x2
	turtle.y = y2

	onStateChanged()

	return nil, nil
}

func _t_Back(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	x2 := int64(math.Cos(delta*turtle.d*dToR)) - turtle.x
	y2 := int64(math.Sin(delta*turtle.d*dToR)) - turtle.y

	if turtle.penState == penDown {
		onDrawLine(turtle.x, turtle.y, x2, y2)
	}

	turtle.x = x2
	turtle.y = y2

	onStateChanged()

	return nil, nil
}

func _t_Left(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	turtle.d -= delta
	for turtle.d < 0 {
		turtle.d += 360
	}

	onStateChanged()
	return nil, nil
}

func _t_Right(frame Frame, parameters []Node) (Node, error) {

	delta, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	turtle.d += delta
	for turtle.d >= 360 {
		turtle.d -= 360
	}

	onStateChanged()
	return nil, nil
}
