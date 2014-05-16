package main

import (
	"io"
	"math"
	"math/rand"
	"strings"
	"time"
)

var keywordTo string = "TO"
var keywordEnd string = "END"
var keywordTrue string = "TRUE"
var keywordFalse string = "FALSE"
var keywordThing string = "THING"

var trueNode Node = newWordNode(-1, -1, keywordTrue, true)
var falseNode Node = newWordNode(-1, -1, keywordFalse, true)
var randomMax Node = newWordNode(-1, -1, "10", true)

var traceEnabled bool

func _bi_Output(frame Frame, parameters []Node) (Node, error) {

	f, err := findInterpretedFrame(frame)
	if err != nil {
		return nil, err
	}
	f.setReturnValue(parameters[0])
	return nil, nil
}

func _bi_Stop(frame Frame, parameters []Node) (Node, error) {
	f, err := findInterpretedFrame(frame)
	if err != nil {
		return nil, err
	}
	f.stop()
	return nil, nil
}

func _bi_Repeat(frame Frame, parameters []Node) (Node, error) {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}
	nn := int64(n)

	for ix := int64(0); ix < nn; ix++ {
		_, err = evalInstructionList(frame, parameters[1], false)
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func _bi_Print(frame Frame, parameters []Node) (Node, error) {

	printNode(frame.workspace(), parameters[0], false)
	frame.workspace().print("\n")
	return nil, nil
}

func _bi_FPrint(frame Frame, parameters []Node) (Node, error) {

	printNode(frame.workspace(), parameters[0], true)
	frame.workspace().print("\n")
	return nil, nil
}

func _bi_Type(frame Frame, parameters []Node) (Node, error) {

	printNode(frame.workspace(), parameters[0], false)
	return nil, nil
}

func _bi_FType(frame Frame, parameters []Node) (Node, error) {

	printNode(frame.workspace(), parameters[0], true)
	return nil, nil
}

func _bi_If(frame Frame, parameters []Node) (Node, error) {

	r, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}

	if r {
		return evalInstructionList(frame, parameters[1], true)
	}

	return nil, nil
}

func _bi_ReadList(frame Frame, parameters []Node) (Node, error) {

	line, err := frame.workspace().files.reader.ReadLine()
	if err != nil {
		return nil, err
	}
	return ParseString("[ " + line + " ]")
}

func _bi_Request(frame Frame, parameters []Node) (Node, error) {

	fw := frame.workspace().files.reader
	fr := frame.workspace().files.writer
	fw.Write(promptPrimary)
	line, err := fr.ReadLine()
	if err != nil {
		return nil, err
	}
	return ParseString("[ " + line + " ]")
}

func _bi_IfElse(frame Frame, parameters []Node) (Node, error) {

	r, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}

	if r {
		return evalInstructionList(frame, parameters[1], true)
	} else {
		return evalInstructionList(frame, parameters[2], true)
	}
}

func _bi_Sum(frame Frame, parameters []Node) (Node, error) {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	return createNumericNode(x + y), nil
}

func _bi_Difference(frame Frame, parameters []Node) (Node, error) {

	var x, y float64
	var err error
	x, y, err = evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	if frame.caller().isInfix {
		return createNumericNode(y - x), nil
	}
	return createNumericNode(x - y), nil
}

func _bi_Product(frame Frame, parameters []Node) (Node, error) {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	return createNumericNode(x * y), nil
}

func _bi_Quotient(frame Frame, parameters []Node) (Node, error) {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	if frame.caller().isInfix {
		return createNumericNode(y / x), nil
	}
	return createNumericNode(x / y), nil
}

func _bi_Remainder(frame Frame, parameters []Node) (Node, error) {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	return createNumericNode(float64(int64(x) % int64(y))), nil
}

func _bi_Maximum(frame Frame, parameters []Node) (Node, error) {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	return createNumericNode(math.Max(x, y)), nil
}

func _bi_Minimum(frame Frame, parameters []Node) (Node, error) {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
	}

	return createNumericNode(math.Min(x, y)), nil
}

func _bi_Equalp(frame Frame, parameters []Node) (Node, error) {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y, true) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Is(frame Frame, parameters []Node) (Node, error) {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y, false) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_NotEqualp(frame Frame, parameters []Node) (Node, error) {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y, true) {
		return falseNode, nil
	}
	return trueNode, nil
}

func _bi_Greaterp(frame Frame, parameters []Node) (Node, error) {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx < ny) || (!infix && nx > ny)) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Lessp(frame Frame, parameters []Node) (Node, error) {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx > ny) || (!infix && nx < ny)) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_GreaterEqualp(frame Frame, parameters []Node) (Node, error) {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx <= ny) || (!infix && nx >= ny)) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_LessEqualp(frame Frame, parameters []Node) (Node, error) {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx >= ny) || (!infix && nx <= ny)) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Numberp(frame Frame, parameters []Node) (Node, error) {

	_, err := evalToNumber(parameters[0])
	if err != nil {
		return falseNode, nil
	}
	return trueNode, nil
}

func _bi_Zerop(frame Frame, parameters []Node) (Node, error) {

	n, err := evalToNumber(parameters[0])
	if err != nil || n != 0.0 {
		return falseNode, nil
	}
	return trueNode, nil
}

func _bi_Rnd(frame Frame, parameters []Node) (Node, error) {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}
	nn := int64(n)
	if nn < 1 {
		return nil, errorPositiveIntegerExpected(parameters[0])
	}

	return createNumericNode(float64(rand.Int63n(nn))), nil
}

func _bi_Random(frame Frame, parameters []Node) (Node, error) {

	return _bi_Rnd(frame, []Node{randomMax})
}

func _bi_Sqrt(frame Frame, parameters []Node) (Node, error) {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	if n <= 0 {
		return nil, errorPositiveNumberExpected(parameters[0])
	}

	return createNumericNode(math.Sqrt(n)), nil
}

func _bi_Pow(frame Frame, parameters []Node) (Node, error) {
	nx, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}
	ny, err := evalToNumber(parameters[1])
	if err != nil {
		return nil, err
	}

	return createNumericNode(math.Pow(nx, ny)), nil
}

func _bi_Sin(frame Frame, parameters []Node) (Node, error) {

	a, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	return createNumericNode(math.Sin(a * (180.0 / math.Pi))), nil
}

func _bi_Cos(frame Frame, parameters []Node) (Node, error) {

	a, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	return createNumericNode(math.Cos(a * (180.0 / math.Pi))), nil
}

func _bi_Arctan(frame Frame, parameters []Node) (Node, error) {

	a, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	return createNumericNode(math.Atan(a * (180.0 / math.Pi))), nil
}

func _bi_Test(frame Frame, parameters []Node) (Node, error) {

	b, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}

	if b {
		frame.parentFrame().setTestValue(trueNode)
	} else {
		frame.parentFrame().setTestValue(falseNode)
	}

	return nil, nil
}

func _bi_IfTrue(frame Frame, parameters []Node) (Node, error) {

	n := frame.parentFrame().getTestValue()
	if n == nil {
		return nil, errorNoCurrentTestValue(parameters[0])
	}

	if nodesEqual(n, trueNode, false) {
		return evalInstructionList(frame, parameters[0], true)
	}
	return nil, nil
}

func _bi_IfFalse(frame Frame, parameters []Node) (Node, error) {

	n := frame.parentFrame().getTestValue()
	if n == nil {
		return nil, errorNoCurrentTestValue(parameters[0])
	}

	if nodesEqual(n, falseNode, false) {
		return evalInstructionList(frame, parameters[0], true)
	}
	return nil, nil
}

func _bi_Make(frame Frame, parameters []Node) (Node, error) {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	frame.parentFrame().setVariable(name, parameters[1])
	return nil, nil
}

func _bi_Thing(frame Frame, parameters []Node) (Node, error) {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	val := frame.parentFrame().getVariable(name)
	if val == nil {
		return nil, errorVariableNotFound(parameters[0], name)
	}

	return val, nil
}

func _bi_Local(frame Frame, parameters []Node) (Node, error) {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	frame.parentFrame().createLocal(name)
	return nil, nil
}

func _bi_Word(frame Frame, parameters []Node) (Node, error) {

	l, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	r, err := evalToWord(parameters[1])
	if err != nil {
		return nil, err
	}

	return newWordNode(-1, -1, l+r, true), nil
}

func _bi_Sentence(frame Frame, parameters []Node) (Node, error) {

	var fn Node

	switch l := parameters[0].(type) {
	case *WordNode:
		fn = l.clone()
	case *ListNode:
		fn = (l.clone()).(*ListNode).firstChild
	}

	nn := fn
	for nn.next() != nil {
		nn = nn.next()
	}

	switch r := parameters[1].(type) {
	case *WordNode:
		nn.addNode(r.clone())
	case *ListNode:
		nn.addNode((r.clone()).(*ListNode).firstChild)
	}

	return newListNode(-1, -1, fn), nil
}

func _bi_List(frame Frame, parameters []Node) (Node, error) {

	l := parameters[0].clone()
	l.addNode(parameters[1].clone())

	return newListNode(-1, -1, l), nil
}

func _bi_FPut(frame Frame, parameters []Node) (Node, error) {

	l := parameters[0].clone()

	switch r := parameters[1].(type) {
	case *ListNode:
		rc := r.clone().(*ListNode)
		l.addNode(rc.firstChild)
		rc.firstChild = l
		return rc, nil
	}
	return nil, errorListExpected(parameters[1])
}

func _bi_LPut(frame Frame, parameters []Node) (Node, error) {

	l := parameters[0].clone()

	switch r := parameters[1].(type) {
	case *ListNode:
		rc := r.clone().(*ListNode)
		n := rc.firstChild
		for n != nil && n.next() != nil {
			n = n.next()
		}
		if n == nil {
			rc.firstChild = l
		} else {
			n.addNode(l)
		}
		return rc, nil
	}
	return nil, errorListExpected(parameters[1])
}

func _bi_First(frame Frame, parameters []Node) (Node, error) {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return nil, errorBadInput(n)
		}
		return newWordNode(-1, -1, string(n.value[0]), true), nil
	case *ListNode:
		if n.firstChild == nil {
			return nil, errorBadInput(n)
		}
		return n.firstChild.clone(), nil

	}

	return nil, nil
}

func _bi_Last(frame Frame, parameters []Node) (Node, error) {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return nil, errorBadInput(n)
		}
		return newWordNode(-1, -1, string(n.value[len(n.value)-1]), true), nil
	case *ListNode:
		if n.firstChild == nil {
			return nil, errorBadInput(n)
		}
		nn := n.firstChild
		for nn.next() != nil {
			nn = nn.next()
		}
		return nn.clone(), nil
	}

	return nil, nil
}

func _bi_ButFirst(frame Frame, parameters []Node) (Node, error) {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return nil, errorBadInput(n)
		}
		return newWordNode(-1, -1, string(n.value[1:]), true), nil
	case *ListNode:
		if n.firstChild == nil {
			return nil, errorBadInput(n)
		}
		nn := n.clone().(*ListNode)
		nn.firstChild = nn.firstChild.next()
		return nn, nil
	}

	return nil, nil
}

func _bi_ButLast(frame Frame, parameters []Node) (Node, error) {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return nil, errorBadInput(n)
		}
		return newWordNode(-1, -1, string(n.value[0:len(n.value)-1]), true), nil
	case *ListNode:
		if n.firstChild == nil {
			return nil, errorBadInput(n)
		}
		nn := n.clone().(*ListNode)
		var pn Node = nil
		for cn := nn.firstChild; cn.next() != nil; cn = cn.next() {
			pn = cn
		}
		if pn != nil {
			pn.addNode(nil)
		}
		return nn, nil
	}

	return nil, nil
}

func _bi_Count(frame Frame, parameters []Node) (Node, error) {

	switch n := parameters[0].(type) {
	case *WordNode:
		return createNumericNode(float64(len(n.value))), nil
	case *ListNode:
		return createNumericNode(float64(n.length())), nil
	}

	return nil, nil
}

func _bi_Emptyp(frame Frame, parameters []Node) (Node, error) {
	var length int = 0

	switch n := parameters[0].(type) {
	case *WordNode:
		length = len(n.value)
	case *ListNode:
		length = n.length()
	}

	if length == 0 {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Wordp(frame Frame, parameters []Node) (Node, error) {

	switch parameters[0].(type) {
	case *WordNode:
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Sentencep(frame Frame, parameters []Node) (Node, error) {

	switch n := parameters[0].(type) {
	case *ListNode:
		for nn := n.firstChild; nn != nil; nn = nn.next() {
			if nn.nodeType() != Word {
				return falseNode, nil
			}
		}
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Memberp(frame Frame, parameters []Node) (Node, error) {

	switch y := parameters[1].(type) {
	case *WordNode:
		if parameters[0].nodeType() != Word {
			return nil, errorBadInput(parameters[0])
		}
		x := parameters[0].(*WordNode)
		if len(x.value) != 1 {
			return nil, errorBadInput(x)
		}

		if strings.Index(strings.ToUpper(y.value), strings.ToUpper(x.value)) >= 0 {
			return trueNode, nil
		}
	case *ListNode:
		for c := y.firstChild; c != nil; c = c.next() {
			if nodesEqual(c, parameters[0], false) {
				return trueNode, nil
			}
		}
	}

	return falseNode, nil
}

func _bi_Item(frame Frame, parameters []Node) (Node, error) {

	ix := int64(0)
	switch n := parameters[0].(type) {
	case *WordNode:
		fix, err := evalToNumber(n)
		if err != nil {
			return nil, err
		}
		ix = int64(fix)

	case *ListNode:
		return nil, errorBadInput(n)
	}

	if ix <= 0 {
		return nil, errorBadInput(parameters[0])
	}

	switch v := parameters[1].(type) {
	case *WordNode:
		if ix > int64(len(v.value)) {
			return nil, errorBadInput(parameters[0])
		}
		return newWordNode(-1, -1, string(v.value[ix-1]), true), nil

	case *ListNode:
		cn := v.firstChild
		for i := int64(1); i < ix; i++ {
			cn = cn.next()
		}
		if cn == nil {
			return nil, errorBadInput(parameters[0])
		}
		return cn.clone(), nil
	}

	return nil, nil
}

func _bi_Goodbye(frame Frame, parameters []Node) (Node, error) {

	frame.workspace().print("Seeya!\n\n")

	frame.workspace().broker.PublishId(MT_Quit)

	return nil, nil
}

func _bi_Both(frame Frame, parameters []Node) (Node, error) {

	x, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}
	if !x {
		return falseNode, nil
	}

	y, err := evalToBoolean(parameters[1])
	if err != nil {
		return nil, err
	}

	if !y {
		return falseNode, nil
	}

	return trueNode, nil
}

func _bi_Either(frame Frame, parameters []Node) (Node, error) {

	x, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}
	if x {
		return trueNode, nil
	}

	y, err := evalToBoolean(parameters[1])
	if err != nil {
		return nil, err
	}

	if y {
		return trueNode, nil
	}

	return falseNode, nil
}

func _bi_Not(frame Frame, parameters []Node) (Node, error) {

	x, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}
	if x {
		return falseNode, nil
	}
	return trueNode, nil
}

func _bi_Trace(frame Frame, parameters []Node) (Node, error) {

	frame.workspace().setTrace(true)

	return nil, nil
}

func _bi_Untrace(frame Frame, parameters []Node) (Node, error) {

	frame.workspace().setTrace(false)

	return nil, nil
}

func _bi_Wait(frame Frame, parameters []Node) (Node, error) {

	fs, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}

	s := time.Duration(fs)
	if s < 1 {
		return nil, errorBadInput(parameters[0])
	}

	time.Sleep(s * time.Second)

	return nil, nil
}

func _bi_Run(frame Frame, parameters []Node) (Node, error) {

	return evalInstructionList(frame, parameters[0], false)
}

func _bi_Po(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}
	ws := frame.workspace()
	for _, n := range names {
		p := ws.findProcedure(strings.ToUpper(n.value))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				ws.print(ip.source)
				ws.print("\n")
			}
		}
	}

	return nil, nil
}

func _bi_PoAll(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()

	_bi_Pops(frame, parameters)

	if len(ws.rootFrame.vars) > 0 {
		ws.print("\n")

		_bi_Pons(frame, parameters)
	}
	return nil, nil
}

func printVariable(ws *Workspace, n string, v Node) {
	ws.print("MAKE \"" + n + " ")
	switch v.(type) {
	case *WordNode:
		ws.print("\"")
	}
	printNode(ws, v, true)
	ws.print("\n")
}

func printTitle(ws *Workspace, p *InterpretedProcedure) {

	ws.print("TO ")
	ws.print(p.name)

	for _, i := range p.parameters {
		ws.print(" :")
		ws.print(i)
	}
	ws.print("\n")
}

func toWordList(node Node) ([]*WordNode, error) {
	names := make([]*WordNode, 0, 1)
	switch n := node.(type) {
	case *WordNode:
		names = append(names, n)
	case *ListNode:
		for nn := n.firstChild; nn != nil; nn = nn.next() {
			switch wn := nn.(type) {
			case *WordNode:
				names = append(names, wn)
			case *ListNode:
				return nil, errorWordExpected(wn)
			}
		}
	}
	return names, nil
}

func _bi_Pon(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		v, exists := ws.rootFrame.vars[name]
		if exists && !v.buried {
			printVariable(ws, name, v.value)
		}
	}

	return nil, nil
}

func _bi_Pons(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()

	for n, v := range ws.rootFrame.vars {
		if !v.buried {
			printVariable(ws, n, v.value)
		}
	}

	return nil, nil
}

func _bi_Pops(frame Frame, parameters []Node) (Node, error) {
	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				ws.print(ip.source)
				ws.print("\n")
			}
		}
	}

	return nil, nil
}

func _bi_Pots(frame Frame, parameters []Node) (Node, error) {
	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				printTitle(ws, ip)
			}
		}
	}

	return nil, nil
}

func _bi_Pot(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		p := ws.findProcedure(strings.ToUpper(n.value))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				printTitle(ws, ip)
			}
		}
	}

	return nil, nil
}

func _bi_ErAll(frame Frame, parameters []Node) (Node, error) {

	_bi_Erps(frame, parameters)

	_bi_Erns(frame, parameters)

	return nil, nil
}

func _bi_Erase(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		p := ws.findProcedure(strings.ToUpper(name))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				delete(ws.procedures, name)
			}
		}
	}

	return nil, nil
}

func _bi_Ern(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		v, exists := ws.rootFrame.vars[name]
		if exists && !v.buried {
			delete(ws.rootFrame.vars, name)
		}
	}

	return nil, nil
}

func _bi_Erns(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()

	for n, v := range ws.rootFrame.vars {
		if !v.buried {
			delete(ws.rootFrame.vars, n)
		}
	}
	return nil, nil
}

func _bi_Erps(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()

	for n, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				delete(ws.procedures, n)
			}
		}
	}

	return nil, nil
}

func _bi_Bury(frame Frame, parameters []Node) (Node, error) {
	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		p := ws.findProcedure(strings.ToUpper(name))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			ip.buried = true
		}
	}

	return nil, nil
}

func _bi_BuryAll(frame Frame, parameters []Node) (Node, error) {

	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			ip.buried = true
		}
	}

	for _, v := range ws.rootFrame.vars {
		v.buried = true
	}
	return nil, nil
}

func _bi_BuryName(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		v, exists := ws.rootFrame.vars[name]
		if exists {
			v.buried = true
		}
	}

	return nil, nil
}

func _bi_Unbury(frame Frame, parameters []Node) (Node, error) {
	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		p := ws.findProcedure(strings.ToUpper(name))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			ip.buried = false
		}
	}

	return nil, nil
}

func _bi_UnburyAll(frame Frame, parameters []Node) (Node, error) {
	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			ip.buried = false
		}
	}

	for _, v := range ws.rootFrame.vars {
		v.buried = false
	}
	return nil, nil
}

func _bi_UnburyName(frame Frame, parameters []Node) (Node, error) {
	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		v, exists := ws.rootFrame.vars[name]
		if exists {
			v.buried = false
		}
	}

	return nil, nil
}

func _bi_Load(frame Frame, parameters []Node) (Node, error) {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	err = ws.files.OpenFile(name)
	if err != nil {
		return nil, err
	}

	err = ws.files.SetReader(name)
	if err != nil {
		return nil, err
	}
	defer ws.files.CloseFile(name)

	err = ws.readFile()
	if err != nil && err != io.EOF {
		return nil, err
	}

	return nil, nil
}

func _bi_Save(frame Frame, parameters []Node) (Node, error) {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	err = ws.files.OpenFile(name)
	if err != nil {
		return nil, err
	}

	err = ws.files.SetWriter(name)
	if err != nil {
		return nil, err
	}
	defer ws.files.CloseFile(name)

	return _bi_PoAll(frame, parameters)
}

func _bi_Savel(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}

	name, err := evalToWord(parameters[1])
	if err != nil {
		return nil, err
	}

	ws := frame.workspace()
	err = ws.files.OpenFile(name)
	if err != nil {
		return nil, err
	}

	err = ws.files.SetWriter(name)
	if err != nil {
		return nil, err
	}
	defer ws.files.CloseFile(name)

	for _, n := range names {
		p := ws.findProcedure(strings.ToUpper(n.value))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				ws.print(ip.source)
				ws.print("\n")
			}
		}
	}

	return _bi_Pons(frame, parameters)
}

func _bi_Catalog(frame Frame, parameters []Node) (Node, error) {

	frame.workspace().files.Catalog()
	return nil, nil
}

func _bi_Prefix(frame Frame, parameters []Node) (Node, error) {

	return newWordNode(-1, -1, frame.workspace().files.rootPath, true), nil
}

func _bi_SetPrefix(frame Frame, parameters []Node) (Node, error) {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}
	return nil, frame.workspace().files.SetPrefix(p)
}

func _bi_CreateDir(frame Frame, parameters []Node) (Node, error) {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}
	return nil, frame.workspace().files.CreateDir(p)
}

func _bi_EraseFile(frame Frame, parameters []Node) (Node, error) {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}
	return nil, frame.workspace().files.EraseFile(p)
}

func _bi_Filep(frame Frame, parameters []Node) (Node, error) {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}
	if frame.workspace().files.IsFile(p) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Rename(frame Frame, parameters []Node) (Node, error) {

	from, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}
	to, err := evalToWord(parameters[1])
	if err != nil {
		return nil, err
	}
	return nil, frame.workspace().files.Rename(from, to)
}

func _bi_Pofile(frame Frame, parameters []Node) (Node, error) {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return nil, err
	}

	fs := frame.workspace().files

	err = fs.OpenFile(p)
	if err != nil {
		return nil, err
	}
	defer fs.CloseFile(p)

	err = fs.SetReader(p)
	if err != nil {
		return nil, err
	}

	for {
		l, err := fs.reader.ReadLine()

		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		fs.writer.Write(l)
		fs.writer.Write("\n")
	}

	return nil, nil
}

func registerBuiltInProcedures(workspace *Workspace) {

	workspace.registerBuiltIn("OUTPUT", "OP", 1, _bi_Output)
	workspace.registerBuiltIn("STOP", "", 0, _bi_Stop)

	workspace.registerBuiltIn("PRINT", "PR", 1, _bi_Print)
	workspace.registerBuiltIn("FPRINT", "FP", 1, _bi_FPrint)
	workspace.registerBuiltIn("TYPE", "TY", 1, _bi_Type)
	workspace.registerBuiltIn("FTYPE", "FTY", 1, _bi_FType)

	workspace.registerBuiltIn("READLIST", "RL", 0, _bi_ReadList)
	workspace.registerBuiltIn("REQUEST", "", 0, _bi_Request)

	workspace.registerBuiltIn("REPEAT", "", 2, _bi_Repeat)
	workspace.registerBuiltIn("IF", "", 2, _bi_If)
	workspace.registerBuiltIn("IFELSE", "", 3, _bi_IfElse)

	workspace.registerBuiltIn("SUM", "", 2, _bi_Sum)
	workspace.registerBuiltIn("DIFFERENCE", "DIFF", 2, _bi_Difference)
	workspace.registerBuiltIn("PRODUCT", "", 2, _bi_Product)
	workspace.registerBuiltIn("QUOTIENT", "", 2, _bi_Quotient)
	workspace.registerBuiltIn("REMAINDER", "MOD", 2, _bi_Remainder)
	workspace.registerBuiltIn("MAXIMUM", "MAX", 2, _bi_Maximum)
	workspace.registerBuiltIn("MINIMUM", "MIN", 2, _bi_Minimum)

	workspace.registerBuiltIn("EQUALP", "EQUAL?", 2, _bi_Equalp)
	workspace.registerBuiltIn("IS", "EQUAL?", 2, _bi_Is)
	workspace.registerBuiltIn("NOTEQUALP", "NOTEQUAL?", 2, _bi_NotEqualp)
	workspace.registerBuiltIn("GREATERP", "GREATER?", 2, _bi_Greaterp)
	workspace.registerBuiltIn("LESSP", "LESS?", 2, _bi_Lessp)
	workspace.registerBuiltIn("GREATEREQUALP", "GREATERQUAL?", 2, _bi_GreaterEqualp)
	workspace.registerBuiltIn("LESSEQUALP", "LESSEQUAL?", 2, _bi_LessEqualp)
	workspace.registerBuiltIn("NUMBERP", "NUMBER?", 1, _bi_Numberp)
	workspace.registerBuiltIn("ZEROP", "ZERO?", 1, _bi_Zerop)

	workspace.registerBuiltIn("RANDOM", "", 0, _bi_Random)
	workspace.registerBuiltIn("RND", "", 1, _bi_Rnd)
	workspace.registerBuiltIn("SQRT", "", 1, _bi_Sqrt)
	workspace.registerBuiltIn("POW", "", 2, _bi_Pow)
	workspace.registerBuiltIn("SIN", "", 1, _bi_Sin)
	workspace.registerBuiltIn("COS", "", 1, _bi_Cos)
	workspace.registerBuiltIn("ARCTAN", "", 1, _bi_Arctan)

	workspace.registerBuiltIn("TEST", "", 1, _bi_Test)
	workspace.registerBuiltIn("IFTRUE", "", 1, _bi_IfTrue)
	workspace.registerBuiltIn("IFFALSE", "", 1, _bi_IfFalse)

	workspace.registerBuiltIn("MAKE", "", 2, _bi_Make)
	workspace.registerBuiltIn("THING", "", 1, _bi_Thing)
	workspace.registerBuiltIn("LOCAL", "", 1, _bi_Local)

	workspace.registerBuiltIn("WORD", "", 2, _bi_Word)
	workspace.registerBuiltIn("SENTENCE", "SE", 2, _bi_Sentence)
	workspace.registerBuiltIn("LIST", "", 2, _bi_List)
	workspace.registerBuiltIn("FPUT", "", 2, _bi_FPut)
	workspace.registerBuiltIn("LPUT", "", 2, _bi_LPut)
	workspace.registerBuiltIn("FIRST", "", 1, _bi_First)
	workspace.registerBuiltIn("LAST", "", 1, _bi_Last)
	workspace.registerBuiltIn("BUTFIRST", "", 1, _bi_ButFirst)
	workspace.registerBuiltIn("BUTLAST", "", 1, _bi_ButLast)

	workspace.registerBuiltIn("COUNT", "", 1, _bi_Count)
	workspace.registerBuiltIn("EMPTYP", "", 1, _bi_Emptyp)
	workspace.registerBuiltIn("WORDP", "", 1, _bi_Wordp)
	workspace.registerBuiltIn("SENTENCEP", "", 1, _bi_Sentencep)
	workspace.registerBuiltIn("MEMBERP", "", 2, _bi_Memberp)
	workspace.registerBuiltIn("ITEM", "NTH", 2, _bi_Item)

	workspace.registerBuiltIn("BOTH", "AND", 2, _bi_Both)
	workspace.registerBuiltIn("EITHER", "OR", 2, _bi_Either)
	workspace.registerBuiltIn("NOT", "", 1, _bi_Not)

	workspace.registerBuiltIn("TRACE", "", 0, _bi_Trace)
	workspace.registerBuiltIn("UNTRACE", "", 0, _bi_Untrace)
	workspace.registerBuiltIn("GOODBYE", "BYE", 0, _bi_Goodbye)
	workspace.registerBuiltIn("WAIT", "", 1, _bi_Wait)

	workspace.registerBuiltIn("RUN", "", 1, _bi_Run)

	workspace.registerBuiltIn("PO", "", 1, _bi_Po)
	workspace.registerBuiltIn("POALL", "", 0, _bi_PoAll)
	workspace.registerBuiltIn("PON", "", 1, _bi_Pon)
	workspace.registerBuiltIn("PONS", "", 0, _bi_Pons)
	workspace.registerBuiltIn("POPS", "", 0, _bi_Pops)
	workspace.registerBuiltIn("POT", "", 1, _bi_Pot)
	workspace.registerBuiltIn("POTS", "", 0, _bi_Pots)

	workspace.registerBuiltIn("ERALL", "", 0, _bi_ErAll)
	workspace.registerBuiltIn("ERASE", "", 1, _bi_Erase)
	workspace.registerBuiltIn("ERN", "", 1, _bi_Ern)
	workspace.registerBuiltIn("ERNS", "", 0, _bi_Erns)
	workspace.registerBuiltIn("ERPS", "", 0, _bi_Erps)

	workspace.registerBuiltIn("BURY", "", 1, _bi_Bury)
	workspace.registerBuiltIn("BURYALL", "", 0, _bi_BuryAll)
	workspace.registerBuiltIn("BURYNAME", "", 1, _bi_BuryName)
	workspace.registerBuiltIn("UNBURY", "", 1, _bi_Unbury)
	workspace.registerBuiltIn("UNBURYALL", "", 0, _bi_UnburyAll)
	workspace.registerBuiltIn("UNBURYNAME", "", 1, _bi_UnburyName)

	workspace.registerBuiltIn("LOAD", "", 1, _bi_Load)
	workspace.registerBuiltIn("SAVE", "", 1, _bi_Save)
	workspace.registerBuiltIn("SAVEL", "", 2, _bi_Savel)

	workspace.registerBuiltIn("CATALOG", "", 0, _bi_Catalog)
	workspace.registerBuiltIn("PREFIX", "", 0, _bi_Prefix)
	workspace.registerBuiltIn("SETPREFIX", "", 1, _bi_SetPrefix)
	workspace.registerBuiltIn("CREATEDIR", "", 1, _bi_CreateDir)
	workspace.registerBuiltIn("ERASEFILE", "", 1, _bi_EraseFile)
	workspace.registerBuiltIn("FILEP", "", 1, _bi_Filep)
	workspace.registerBuiltIn("RENAME", "", 2, _bi_Rename)
	workspace.registerBuiltIn("POFILE", "", 1, _bi_Pofile)

}
