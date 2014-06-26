package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
)

var keywordTo string = "TO"
var keywordEnd string = "END"
var keywordTrue string = "TRUE"
var keywordFalse string = "FALSE"
var keywordThing string = "THING"
var keywordIf string = "IF"
var keywordGo string = "GO"
var keywordLabel string = "LABEL"
var keywordError string = "ERROR"
var keywordEdit string = "EDIT"

var trueNode Node = newWordNode(-1, -1, keywordTrue, true)
var falseNode Node = newWordNode(-1, -1, keywordFalse, true)
var randomMax Node = newWordNode(-1, -1, "10", true)

var traceEnabled bool

func _bi_Go(frame Frame, parameters []Node) *CallResult {
	// Dummy - call handled in eval.go
	return nil
}

func _bi_Label(frame Frame, parameters []Node) *CallResult {
	// Dummy - call handled in eval.go
	return nil
}

func _bi_Output(frame Frame, parameters []Node) *CallResult {

	f, err := findInterpretedFrame(frame)
	if err != nil {
		return errorResult(err)
	}
	f.setReturnValue(parameters[0])
	return stopResult()
}

func _bi_Stop(frame Frame, parameters []Node) *CallResult {
	f, err := findInterpretedFrame(frame)
	if err != nil {
		return errorResult(err)
	}

	f.stop()
	return stopResult()
}

func _bi_Repeat(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	nn := int64(n)

	for ix := int64(0); ix < nn; ix++ {
		cr := evalInstructionList(frame, parameters[1], false)
		if cr != nil && cr.shouldStop() {
			return cr
		}
	}

	return nil
}

func _bi_Print(frame Frame, parameters []Node) *CallResult {

	for _, p := range parameters {
		printNode(frame.workspace(), p, false)
	}
	frame.workspace().print("\n")
	return nil
}

func _bi_FPrint(frame Frame, parameters []Node) *CallResult {

	for _, p := range parameters {
		printNode(frame.workspace(), p, true)
	}
	frame.workspace().print("\n")
	return nil
}

func _bi_Type(frame Frame, parameters []Node) *CallResult {

	for _, p := range parameters {
		printNode(frame.workspace(), p, false)
	}
	return nil
}

func _bi_If(frame Frame, parameters []Node) *CallResult {

	r, err := evalToBoolean(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	if r {
		return evalInstructionList(frame, parameters[1], true)
	}

	if len(parameters) == 3 {
		return evalInstructionList(frame, parameters[2], true)
	}

	return nil
}

func _bi_ReadList(frame Frame, parameters []Node) *CallResult {

	line, err := frame.workspace().files.reader.ReadLine()
	if err != nil {
		return errorResult(err)
	}
	n, err := ParseString("[ " + line + " ]")
	if err != nil {
		return errorResult(err)
	}
	n.setLiteral()
	return returnResult(n)
}

func _bi_ReadWord(frame Frame, parameters []Node) *CallResult {

	line, err := frame.workspace().files.reader.ReadLine()
	if err != nil {
		return errorResult(err)
	}
	return returnResult(newWordNode(-1, -1, line, true))
}

func _bi_ReadChar(frame Frame, parameters []Node) *CallResult {

	c, err := frame.workspace().files.reader.ReadChar()
	if err == io.EOF {
		return returnResult(newListNode(-1, -1, nil))
	}
	return returnResult(newWordNode(-1, -1, string(c), true))
}

func _bi_ReadChars(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	if n < 0 {
		return errorResult(errorBadInput(parameters[0]))
	}

	nn := int(n)
	chars := make([]rune, 0, nn)
	for ix := 0; ix < nn; ix++ {
		c, err := frame.workspace().files.reader.ReadChar()
		if err != nil {
			if err == io.EOF {
				break
			}
			return errorResult(err)
		}
		chars = append(chars, c)
	}

	return returnResult(newWordNode(-1, -1, string(chars), true))
}

func _bi_Request(frame Frame, parameters []Node) *CallResult {

	fw := frame.workspace().files.reader
	fr := frame.workspace().files.writer
	fw.Write(promptPrimary)
	line, err := fr.ReadLine()
	if err != nil {
		return errorResult(err)
	}
	n, err := ParseString("[ " + line + " ]")
	if err != nil {
		return errorResult(err)
	}
	n.setLiteral()
	return returnResult(n)
}

func _bi_Sum(frame Frame, parameters []Node) *CallResult {

	n1, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	for ix := 1; ix < len(parameters); ix++ {
		n2, err := evalToNumber(parameters[ix])
		if err != nil {
			return errorResult(err)
		}
		n1 += n2
	}

	return returnResult(createNumericNode(n1))
}

func _bi_Difference(frame Frame, parameters []Node) *CallResult {

	var x, y float64
	var err error
	x, y, err = evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	if frame.caller().isInfix {
		return returnResult(createNumericNode(y - x))
	}
	return returnResult(createNumericNode(x - y))
}

func _bi_Product(frame Frame, parameters []Node) *CallResult {

	n1, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	for ix := 1; ix < len(parameters); ix++ {
		n2, err := evalToNumber(parameters[ix])
		if err != nil {
			return errorResult(err)
		}
		n1 *= n2
	}

	return returnResult(createNumericNode(n1))
}

func _bi_Quotient(frame Frame, parameters []Node) *CallResult {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	if frame.caller().isInfix {
		if x == 0 {
			return errorResult(errorAttemptToDivideByZero(parameters[0]))
		}
		return returnResult(createNumericNode(y / x))
	}

	if y == 0 {
		return errorResult(errorAttemptToDivideByZero(parameters[1]))
	}
	return returnResult(createNumericNode(x / y))
}

func _bi_IntQuotient(frame Frame, parameters []Node) *CallResult {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	if y == 0 {
		return errorResult(errorAttemptToDivideByZero(parameters[1]))
	}
	return returnResult(createNumericNode(float64(int64(x) / int64(y))))
}

func _bi_Int(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(float64(int64(n))))
}

func _bi_Round(frame Frame, parameters []Node) *CallResult {

	n, prec, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	var r float64
	pow := math.Pow(10, float64(prec))
	i := n * pow
	_, frac := math.Modf(i)
	if frac >= 0.5 {
		r = math.Ceil(i)
	} else {
		r = math.Floor(i)
	}

	return returnResult(createNumericNode(r / pow))
}

func _bi_Form(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	f, err := evalToNumber(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	p, err := evalToNumber(parameters[2])
	if err != nil {
		return errorResult(err)
	}

	fs := fmt.Sprintf("%%%d.%df", int64(f), int64(p))
	return returnResult(newWordNode(-1, -1, fmt.Sprintf(fs, n), true))
}

func _bi_Remainder(frame Frame, parameters []Node) *CallResult {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	if y == 0 {
		return errorResult(errorAttemptToDivideByZero(parameters[1]))
	}

	return returnResult(createNumericNode(float64(int64(x) % int64(y))))
}

func _bi_Maximum(frame Frame, parameters []Node) *CallResult {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(math.Max(x, y)))
}

func _bi_Minimum(frame Frame, parameters []Node) *CallResult {

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(math.Min(x, y)))
}

func _bi_Equalp(frame Frame, parameters []Node) *CallResult {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y, true) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Is(frame Frame, parameters []Node) *CallResult {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y, false) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_NotEqualp(frame Frame, parameters []Node) *CallResult {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y, true) {
		return returnResult(falseNode)
	}
	return returnResult(trueNode)
}

func _bi_Greaterp(frame Frame, parameters []Node) *CallResult {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx < ny) || (!infix && nx > ny)) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Lessp(frame Frame, parameters []Node) *CallResult {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx > ny) || (!infix && nx < ny)) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_GreaterEqualp(frame Frame, parameters []Node) *CallResult {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx <= ny) || (!infix && nx >= ny)) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_LessEqualp(frame Frame, parameters []Node) *CallResult {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	infix := frame.caller().isInfix

	if ex == nil && ey == nil && ((infix && nx >= ny) || (!infix && nx <= ny)) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Numberp(frame Frame, parameters []Node) *CallResult {

	_, err := evalToNumber(parameters[0])
	if err != nil {
		return returnResult(falseNode)
	}
	return returnResult(trueNode)
}

func _bi_Zerop(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil || n != 0.0 {
		return returnResult(falseNode)
	}
	return returnResult(trueNode)
}

func _bi_Rnd(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	nn := int64(n)
	if nn < 1 {
		return errorResult(errorPositiveIntegerExpected(parameters[0]))
	}

	return returnResult(createNumericNode(float64(rand.Int63n(nn))))
}

func _bi_Random(frame Frame, parameters []Node) *CallResult {

	return _bi_Rnd(frame, []Node{randomMax})
}

func _bi_Sqrt(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	if n <= 0 {
		return errorResult(errorPositiveNumberExpected(parameters[0]))
	}

	return returnResult(createNumericNode(math.Sqrt(n)))
}

func _bi_Pow(frame Frame, parameters []Node) *CallResult {
	nx, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	ny, err := evalToNumber(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(math.Pow(nx, ny)))
}

func _bi_Sin(frame Frame, parameters []Node) *CallResult {

	a, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(math.Sin(a * (180.0 / math.Pi))))
}

func _bi_Cos(frame Frame, parameters []Node) *CallResult {

	a, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(math.Cos(a * (180.0 / math.Pi))))
}

func _bi_Arctan(frame Frame, parameters []Node) *CallResult {

	a, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(createNumericNode(math.Atan(a * (180.0 / math.Pi))))
}

func _bi_Test(frame Frame, parameters []Node) *CallResult {

	b, err := evalToBoolean(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	if b {
		frame.parentFrame().setTestValue(trueNode)
	} else {
		frame.parentFrame().setTestValue(falseNode)
	}

	return nil
}

func _bi_IfTrue(frame Frame, parameters []Node) *CallResult {

	n := frame.parentFrame().getTestValue()
	if n == nil {
		return errorResult(errorNoCurrentTestValue(parameters[0]))
	}

	if nodesEqual(n, trueNode, false) {
		return evalInstructionList(frame, parameters[0], true)
	}
	return nil
}

func _bi_IfFalse(frame Frame, parameters []Node) *CallResult {

	n := frame.parentFrame().getTestValue()
	if n == nil {
		return errorResult(errorNoCurrentTestValue(parameters[0]))
	}

	if nodesEqual(n, falseNode, false) {
		return evalInstructionList(frame, parameters[0], true)
	}
	return nil
}

func _bi_Make(frame Frame, parameters []Node) *CallResult {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	frame.getVars().setVariable(frame, name, parameters[1])
	return nil
}

func _bi_Name(frame Frame, parameters []Node) *CallResult {

	name, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	frame.getVars().setVariable(frame, name, parameters[0])
	return nil
}

func _bi_Namep(frame Frame, parameters []Node) *CallResult {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	val := frame.getVars().getVariable(frame, name)
	if val == nil {
		return returnResult(falseNode)
	}

	return returnResult(trueNode)
}

func _bi_Thing(frame Frame, parameters []Node) *CallResult {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	val := frame.getVars().getVariable(frame, name)
	if val == nil {
		return errorResult(errorVariableNotFound(parameters[0], name))
	}

	return returnResult(val)
}

func _bi_Local(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	for _, name := range names {
		frame.parentFrame().getVars().createLocal(name.value)
	}
	return nil
}

func _bi_Erprops(frame Frame, parameters []Node) *CallResult {

	frame.workspace().rootFrame.getVars().clearProps()

	return nil
}

func _bi_Gprop(frame Frame, parameters []Node) *CallResult {

	vName, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	pName, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	p := frame.getVars().getProp(frame, vName, strings.ToUpper(pName))
	if p == nil {
		return returnResult(newListNode(-1, -1, nil))
	}

	return returnResult(p)
}

func _bi_Plist(frame Frame, parameters []Node) *CallResult {

	vName, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	vars := frame.getVars()

	v := vars.getVariableInner(frame, vName, strings.ToUpper(vName), false)
	if v == nil || v.props == nil {
		return returnResult(newListNode(-1, -1, nil))
	}

	var fn Node
	var pn Node
	var n Node
	for k, v := range v.props {
		n = newWordNode(-1, -1, k, true)
		if fn == nil {
			fn = n
		} else {
			pn.addNode(n)
		}

		n.addNode(v)

		pn = v
	}

	return returnResult(newListNode(-1, -1, fn))
}

func _bi_Pprop(frame Frame, parameters []Node) *CallResult {

	vName, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	vars := frame.getVars()

	pName, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	vars.setProp(frame, vName, strings.ToUpper(pName), parameters[2])

	return nil
}

func _bi_Pps(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()
	vs := ws.rootFrame.getVars()
	if vs.vars == nil {
		return nil
	}

	for _, v := range vs.vars {
		if v.hasProps() {
			for pName, pVal := range v.props {
				quote := ""
				if pVal.nodeType() == Word {
					quote = "\""
				}
				ws.print(fmt.Sprintf("PPROP \"%s \"%s %s%s\n",
					v.name, pName, quote, pVal.String()))
			}
		}
	}

	return nil
}

func _bi_Remprop(frame Frame, parameters []Node) *CallResult {

	vName, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	vars := frame.getVars()

	pName, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	vars.clearProp(frame, vName, pName)

	return nil
}

func _bi_Word(frame Frame, parameters []Node) *CallResult {

	l, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	r, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(newWordNode(-1, -1, l+r, true))
}

func _bi_Sentence(frame Frame, parameters []Node) *CallResult {

	var fn Node
	var ln Node
	var c Node
	for _, n := range parameters {

		switch nn := n.(type) {
		case *WordNode:
			c = nn.clone()
		case *ListNode:
			c = (nn.clone()).(*ListNode).firstChild
		}
		if fn == nil {
			fn = c
		} else {
			ln.addNode(c)
		}
		ln = c
	}

	return returnResult(newListNode(-1, -1, fn))
}

func _bi_List(frame Frame, parameters []Node) *CallResult {

	l := parameters[0].clone()
	l.addNode(parameters[1].clone())

	return returnResult(newListNode(-1, -1, l))
}

func _bi_FPut(frame Frame, parameters []Node) *CallResult {

	l := parameters[0].clone()

	switch r := parameters[1].(type) {
	case *ListNode:
		rc := r.clone().(*ListNode)
		l.addNode(rc.firstChild)
		rc.firstChild = l
		return returnResult(rc)
	}
	return errorResult(errorListExpected(parameters[1]))
}

func _bi_LPut(frame Frame, parameters []Node) *CallResult {

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
		return returnResult(rc)
	}
	return errorResult(errorListExpected(parameters[1]))
}

func _bi_Parse(frame Frame, parameters []Node) *CallResult {

	val, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	r := strings.NewReader(val + string(listEnd))
	rr := bufio.NewReader(r)
	l := 0
	c := 0
	fn, err := readUntil(rr, &l, &c, listEnd)
	if err != nil {
		return errorResult(err)
	}

	ln := newListNode(-1, -1, fn)
	ln.setLiteral()

	return returnResult(ln)
}

func _bi_First(frame Frame, parameters []Node) *CallResult {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return errorResult(errorBadInput(n))
		}
		return returnResult(newWordNode(-1, -1, string(n.value[0]), true))
	case *ListNode:
		if n.firstChild == nil {
			return errorResult(errorBadInput(n))
		}
		return returnResult(n.firstChild.clone())

	}

	return nil
}

func _bi_Last(frame Frame, parameters []Node) *CallResult {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return errorResult(errorBadInput(n))
		}
		return returnResult(newWordNode(-1, -1, string(n.value[len(n.value)-1]), true))
	case *ListNode:
		if n.firstChild == nil {
			return errorResult(errorBadInput(n))
		}
		nn := n.firstChild
		for nn.next() != nil {
			nn = nn.next()
		}
		return returnResult(nn.clone())
	}

	return nil
}

func _bi_ButFirst(frame Frame, parameters []Node) *CallResult {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return errorResult(errorBadInput(n))
		}
		return returnResult(newWordNode(-1, -1, string(n.value[1:]), true))
	case *ListNode:
		if n.firstChild == nil {
			return errorResult(errorBadInput(n))
		}
		nn := n.clone().(*ListNode)
		nn.firstChild = nn.firstChild.next()
		return returnResult(nn)
	}

	return nil
}

func _bi_ButLast(frame Frame, parameters []Node) *CallResult {

	switch n := parameters[0].(type) {
	case *WordNode:
		if len(n.value) == 0 {
			return errorResult(errorBadInput(n))
		}
		return returnResult(newWordNode(-1, -1, string(n.value[0:len(n.value)-1]), true))
	case *ListNode:
		if n.firstChild == nil {
			return errorResult(errorBadInput(n))
		}
		nn := n.clone().(*ListNode)
		var pn Node = nil
		for cn := nn.firstChild; cn.next() != nil; cn = cn.next() {
			pn = cn
		}
		if pn != nil {
			pn.addNode(nil)
		}
		return returnResult(nn)
	}

	return nil
}

func _bi_Count(frame Frame, parameters []Node) *CallResult {

	switch n := parameters[0].(type) {
	case *WordNode:
		return returnResult(createNumericNode(float64(len(n.value))))
	case *ListNode:
		return returnResult(createNumericNode(float64(n.length())))
	}

	return nil
}

func _bi_Emptyp(frame Frame, parameters []Node) *CallResult {
	var length int = 0

	switch n := parameters[0].(type) {
	case *WordNode:
		length = len(n.value)
	case *ListNode:
		length = n.length()
	}

	if length == 0 {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Wordp(frame Frame, parameters []Node) *CallResult {

	switch parameters[0].(type) {
	case *WordNode:
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Listp(frame Frame, parameters []Node) *CallResult {

	switch parameters[0].(type) {
	case *ListNode:
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Sentencep(frame Frame, parameters []Node) *CallResult {

	switch n := parameters[0].(type) {
	case *ListNode:
		for nn := n.firstChild; nn != nil; nn = nn.next() {
			if nn.nodeType() != Word {
				return returnResult(falseNode)
			}
		}
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Memberp(frame Frame, parameters []Node) *CallResult {

	switch y := parameters[1].(type) {
	case *WordNode:
		if parameters[0].nodeType() != Word {
			return errorResult(errorBadInput(parameters[0]))
		}
		x := parameters[0].(*WordNode)
		if len(x.value) != 1 {
			return errorResult(errorBadInput(x))
		}

		if strings.Index(strings.ToUpper(y.value), strings.ToUpper(x.value)) >= 0 {
			return returnResult(trueNode)
		}
	case *ListNode:
		for c := y.firstChild; c != nil; c = c.next() {
			if nodesEqual(c, parameters[0], false) {
				return returnResult(trueNode)
			}
		}
	}

	return returnResult(falseNode)
}

func _bi_Item(frame Frame, parameters []Node) *CallResult {

	ix := int64(0)
	switch n := parameters[0].(type) {
	case *WordNode:
		fix, err := evalToNumber(n)
		if err != nil {
			return errorResult(err)
		}
		ix = int64(fix)

	case *ListNode:
		return errorResult(errorBadInput(n))
	}

	if ix <= 0 {
		return errorResult(errorBadInput(parameters[0]))
	}

	switch v := parameters[1].(type) {
	case *WordNode:
		if ix > int64(len(v.value)) {
			return errorResult(errorBadInput(parameters[0]))
		}
		return returnResult(newWordNode(-1, -1, string(v.value[ix-1]), true))

	case *ListNode:
		cn := v.firstChild
		for i := int64(1); i < ix; i++ {
			cn = cn.next()
		}
		if cn == nil {
			return errorResult(errorBadInput(parameters[0]))
		}
		return returnResult(cn.clone())
	}

	return nil
}

func _bi_Goodbye(frame Frame, parameters []Node) *CallResult {

	frame.workspace().print("Seeya!\n\n")

	frame.workspace().broker.PublishId(MT_Quit)

	return nil
}

func _bi_Both(frame Frame, parameters []Node) *CallResult {

	for _, v := range parameters {
		x, err := evalToBoolean(v)
		if err != nil {
			return errorResult(err)
		}
		if !x {
			return returnResult(falseNode)
		}
	}

	return returnResult(trueNode)
}

func _bi_Either(frame Frame, parameters []Node) *CallResult {

	for _, v := range parameters {
		x, err := evalToBoolean(v)
		if err != nil {
			return errorResult(err)
		}
		if x {
			return returnResult(trueNode)
		}
	}

	return returnResult(falseNode)
}

func _bi_Not(frame Frame, parameters []Node) *CallResult {

	x, err := evalToBoolean(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	if x {
		return returnResult(falseNode)
	}
	return returnResult(trueNode)
}

func _bi_Trace(frame Frame, parameters []Node) *CallResult {

	frame.workspace().setTrace(true)

	return nil
}

func _bi_Untrace(frame Frame, parameters []Node) *CallResult {

	frame.workspace().setTrace(false)

	return nil
}

func _bi_Wait(frame Frame, parameters []Node) *CallResult {

	fs, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	s := time.Duration(fs)
	if s < 1 {
		return errorResult(errorBadInput(parameters[0]))
	}

	time.Sleep(s * time.Second)

	return nil
}

func _bi_Run(frame Frame, parameters []Node) *CallResult {

	return evalInstructionList(frame, parameters[0], false)
}

func _bi_Po(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
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

	return nil
}

func _bi_PoAll(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()

	_bi_Pops(frame, parameters)

	if len(ws.rootFrame.getVars().vars) > 0 {
		ws.print("\n")

		_bi_Pons(frame, parameters)
	}
	return nil
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

func _bi_Pon(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		v, exists := ws.rootFrame.getVars().vars[name]
		if exists && !v.buried {
			printVariable(ws, name, v.value)
		}
	}

	return nil
}

func _bi_Pons(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()

	for n, v := range ws.rootFrame.getVars().vars {
		if !v.buried {
			printVariable(ws, n, v.value)
		}
	}

	return nil
}

func _bi_Pops(frame Frame, parameters []Node) *CallResult {
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

	return nil
}

func _bi_Pots(frame Frame, parameters []Node) *CallResult {
	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				printTitle(ws, ip)
			}
		}
	}

	return nil
}

func _bi_Pot(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
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

	return nil
}

func _bi_ErAll(frame Frame, parameters []Node) *CallResult {

	_bi_Erps(frame, parameters)

	_bi_Erns(frame, parameters)

	return nil
}

func _bi_Erase(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
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

	return nil
}

func _bi_Ern(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	for _, n := range names {
		name := strings.ToUpper(n.value)
		v, exists := ws.rootFrame.getVars().vars[name]
		if exists && !v.buried {
			delete(ws.rootFrame.getVars().vars, name)
		}
	}

	return nil
}

func _bi_Erns(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()

	for n, v := range ws.rootFrame.getVars().vars {
		if !v.buried {
			delete(ws.rootFrame.getVars().vars, n)
		}
	}
	return nil
}

func _bi_Erps(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()

	for n, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			if !ip.buried {
				delete(ws.procedures, n)
			}
		}
	}

	return nil
}

func _bi_Bury(frame Frame, parameters []Node) *CallResult {
	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
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

	return nil
}

func _bi_BuryAll(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			ip.buried = true
		}
	}

	ws.rootFrame.getVars().setAllBuried(ws.rootFrame, true)

	return nil
}

func _bi_BuryName(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	for _, n := range names {
		ws.rootFrame.getVars().setBuried(ws.rootFrame, n.value, true)
	}

	return nil
}

func _bi_Unbury(frame Frame, parameters []Node) *CallResult {
	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
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

	return nil
}

func _bi_UnburyAll(frame Frame, parameters []Node) *CallResult {
	ws := frame.workspace()

	for _, p := range ws.procedures {
		switch ip := p.(type) {
		case *InterpretedProcedure:
			ip.buried = false
		}
	}

	ws.rootFrame.getVars().setAllBuried(ws.rootFrame, false)

	return nil
}

func _bi_UnburyName(frame Frame, parameters []Node) *CallResult {
	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	for _, n := range names {
		ws.rootFrame.getVars().setBuried(ws.rootFrame, n.value, false)
	}

	return nil
}

func _bi_Load(frame Frame, parameters []Node) *CallResult {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	err = ws.files.OpenFile(name)
	if err != nil {
		return errorResult(err)
	}

	err = ws.files.SetReader(name)
	if err != nil {
		return errorResult(err)
	}
	defer ws.files.CloseFile(name)

	err = ws.readFile()
	if err != nil && err != io.EOF {
		return errorResult(err)
	}

	return nil
}

func _bi_Save(frame Frame, parameters []Node) *CallResult {

	name, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	err = ws.files.OpenFile(name)
	if err != nil {
		return errorResult(err)
	}

	err = ws.files.SetWriter(name)
	if err != nil {
		return errorResult(err)
	}
	defer ws.files.CloseFile(name)

	return _bi_PoAll(frame, parameters)
}

func _bi_Savel(frame Frame, parameters []Node) *CallResult {

	names, err := toWordList(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	name, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}

	ws := frame.workspace()
	err = ws.files.OpenFile(name)
	if err != nil {
		return errorResult(err)
	}

	err = ws.files.SetWriter(name)
	if err != nil {
		return errorResult(err)
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

func _bi_Catalog(frame Frame, parameters []Node) *CallResult {

	frame.workspace().files.Catalog()
	return nil
}

func _bi_Prefix(frame Frame, parameters []Node) *CallResult {

	return returnResult(newWordNode(-1, -1, frame.workspace().files.rootPath, true))
}

func _bi_SetPrefix(frame Frame, parameters []Node) *CallResult {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	err = frame.workspace().files.SetPrefix(p)
	if err != nil {
		return errorResult(err)
	}
	return nil
}

func _bi_CreateDir(frame Frame, parameters []Node) *CallResult {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	err = frame.workspace().files.CreateDir(p)
	if err != nil {
		return errorResult(err)
	}
	return nil
}

func _bi_EraseFile(frame Frame, parameters []Node) *CallResult {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	err = frame.workspace().files.EraseFile(p)
	if err != nil {
		return errorResult(err)
	}
	return nil
}

func _bi_Filep(frame Frame, parameters []Node) *CallResult {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	if frame.workspace().files.IsFile(p) {
		return returnResult(trueNode)
	}
	return returnResult(falseNode)
}

func _bi_Rename(frame Frame, parameters []Node) *CallResult {

	from, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	to, err := evalToWord(parameters[1])
	if err != nil {
		return errorResult(err)
	}
	err = frame.workspace().files.Rename(from, to)
	if err != nil {
		return errorResult(err)
	}
	return nil
}

func _bi_Pofile(frame Frame, parameters []Node) *CallResult {

	p, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	fs := frame.workspace().files

	err = fs.OpenFile(p)
	if err != nil {
		return errorResult(err)
	}
	defer fs.CloseFile(p)

	err = fs.SetReader(p)
	if err != nil {
		return errorResult(err)
	}

	for {
		l, err := fs.reader.ReadLine()

		if err != nil {
			if err == io.EOF {
				break
			}
			return errorResult(err)
		}

		fs.writer.Write(l)
		fs.writer.Write("\n")
	}

	return nil
}

func _bi_Ascii(frame Frame, parameters []Node) *CallResult {

	v, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	if len(v) == 0 {
		return errorResult(errorAtLeastOneCharExpected(parameters[0]))
	}

	b := v[0]
	if b < 128 {
		return returnResult(newWordNode(-1, -1, fmt.Sprint(b), true))
	}
	return returnResult(newWordNode(-1, -1, "-1", true))
}

func _bi_Unicode(frame Frame, parameters []Node) *CallResult {

	v, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	if len(v) == 0 {
		return errorResult(errorAtLeastOneCharExpected(parameters[0]))
	}

	r, _ := utf8.DecodeRuneInString(v)
	if r == utf8.RuneError {
		return errorResult(errorAtLeastOneCharExpected(parameters[0]))
	}

	return returnResult(newWordNode(-1, -1, fmt.Sprint(r), true))
}

func _bi_Char(frame Frame, parameters []Node) *CallResult {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	r := rune(n)
	return returnResult(newWordNode(-1, -1, string(r), true))
}

func _bi_Uppercase(frame Frame, parameters []Node) *CallResult {

	v, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(newWordNode(-1, -1, strings.ToUpper(v), true))
}

func _bi_Lowercase(frame Frame, parameters []Node) *CallResult {

	v, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return returnResult(newWordNode(-1, -1, strings.ToLower(v), true))
}

func _bi_Throw(frame Frame, parameters []Node) *CallResult {

	v, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	return errorResult(userError(v))
}

func _bi_Catch(frame Frame, parameters []Node) *CallResult {

	v, err := evalToWord(parameters[0])
	if err != nil {
		return errorResult(err)
	}

	rv := evalInstructionList(frame, parameters[1], true)
	if !rv.hasError() {
		return rv
	}

	uExpected := strings.ToUpper(v)

	if uExpected == keywordError || uExpected == strings.ToUpper(rv.err.Error()) {
		return nil
	}

	return rv
}

func registerBuiltInProcedures(workspace *Workspace) {

	workspace.registerBuiltIn("OUTPUT", "OP", 1, _bi_Output)
	workspace.registerBuiltIn("STOP", "", 0, _bi_Stop)
	workspace.registerBuiltIn("CATCH", "", 2, _bi_Catch)
	workspace.registerBuiltIn("THROW", "", 1, _bi_Throw)

	workspace.registerBuiltInWithVarParams("PRINT", "PR", 1, _bi_Print)
	workspace.registerBuiltInWithVarParams("SHOW", "", 1, _bi_FPrint)
	workspace.registerBuiltInWithVarParams("TYPE", "TY", 1, _bi_Type)

	workspace.registerBuiltIn("READLIST", "RL", 0, _bi_ReadList)
	workspace.registerBuiltIn("READWORD", "RW", 0, _bi_ReadWord)
	workspace.registerBuiltIn("READCHAR", "RC", 0, _bi_ReadChar)
	workspace.registerBuiltIn("READCHARS", "RCS", 1, _bi_ReadChars)
	workspace.registerBuiltIn("REQUEST", "", 0, _bi_Request)

	workspace.registerBuiltIn("REPEAT", "", 2, _bi_Repeat)
	workspace.registerBuiltIn(keywordIf, "", 2, _bi_If)

	workspace.registerBuiltInWithVarParams("SUM", "", 2, _bi_Sum)
	workspace.registerBuiltIn("DIFFERENCE", "DIFF", 2, _bi_Difference)
	workspace.registerBuiltInWithVarParams("PRODUCT", "", 2, _bi_Product)
	workspace.registerBuiltIn("QUOTIENT", "", 2, _bi_Quotient)
	workspace.registerBuiltIn("INTQUOTIENT", "", 2, _bi_IntQuotient)
	workspace.registerBuiltIn("INT", "", 1, _bi_Int)
	workspace.registerBuiltIn("ROUND", "", 2, _bi_Round)
	workspace.registerBuiltIn("FORM", "", 3, _bi_Form)
	workspace.registerBuiltIn("REMAINDER", "MOD", 2, _bi_Remainder)
	workspace.registerBuiltInWithVarParams("MAXIMUM", "MAX", 2, _bi_Maximum)
	workspace.registerBuiltInWithVarParams("MINIMUM", "MIN", 2, _bi_Minimum)

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
	workspace.registerBuiltIn("IFTRUE", "IFT", 1, _bi_IfTrue)
	workspace.registerBuiltIn("IFFALSE", "IFF", 1, _bi_IfFalse)

	workspace.registerBuiltIn("MAKE", "", 2, _bi_Make)
	workspace.registerBuiltIn("NAME", "", 2, _bi_Name)
	workspace.registerBuiltIn("NAMEP", "", 1, _bi_Namep)
	workspace.registerBuiltIn("THING", "", 1, _bi_Thing)
	workspace.registerBuiltIn("LOCAL", "", 1, _bi_Local)

	workspace.registerBuiltInWithVarParams("WORD", "", 2, _bi_Word)
	workspace.registerBuiltInWithVarParams("SENTENCE", "SE", 2, _bi_Sentence)
	workspace.registerBuiltInWithVarParams("LIST", "", 2, _bi_List)
	workspace.registerBuiltIn("FPUT", "", 2, _bi_FPut)
	workspace.registerBuiltIn("LPUT", "", 2, _bi_LPut)
	workspace.registerBuiltIn("FIRST", "", 1, _bi_First)
	workspace.registerBuiltIn("LAST", "", 1, _bi_Last)
	workspace.registerBuiltIn("BUTFIRST", "BF", 1, _bi_ButFirst)
	workspace.registerBuiltIn("BUTLAST", "BL", 1, _bi_ButLast)
	workspace.registerBuiltIn("PARSE", "", 1, _bi_Parse)
	workspace.registerBuiltIn("ASCII", "", 1, _bi_Ascii)
	workspace.registerBuiltIn("UNICODE", "", 1, _bi_Unicode)
	workspace.registerBuiltIn("CHAR", "", 1, _bi_Char)
	workspace.registerBuiltIn("UPPERCASE", "", 1, _bi_Uppercase)
	workspace.registerBuiltIn("LOWERCASE", "", 1, _bi_Lowercase)

	workspace.registerBuiltIn("COUNT", "", 1, _bi_Count)
	workspace.registerBuiltIn("EMPTYP", "", 1, _bi_Emptyp)
	workspace.registerBuiltIn("WORDP", "", 1, _bi_Wordp)
	workspace.registerBuiltIn("LISTP", "", 1, _bi_Listp)
	workspace.registerBuiltIn("SENTENCEP", "", 1, _bi_Sentencep)
	workspace.registerBuiltIn("MEMBERP", "", 2, _bi_Memberp)
	workspace.registerBuiltIn("ITEM", "NTH", 2, _bi_Item)

	workspace.registerBuiltInWithVarParams("AND", "", 2, _bi_Both)
	workspace.registerBuiltInWithVarParams("OR", "", 2, _bi_Either)
	workspace.registerBuiltInWithVarParams("NOT", "", 1, _bi_Not)

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
	workspace.registerBuiltIn("ERASEFILE", "ERF", 1, _bi_EraseFile)
	workspace.registerBuiltIn("FILEP", "", 1, _bi_Filep)
	workspace.registerBuiltIn("RENAME", "", 2, _bi_Rename)
	workspace.registerBuiltIn("POFILE", "", 1, _bi_Pofile)

	workspace.registerBuiltIn("GO", "", 1, _bi_Go)
	workspace.registerBuiltIn("LABEL", "", 1, _bi_Label)

	workspace.registerBuiltIn("ERPROPS", "", 0, _bi_Erprops)
	workspace.registerBuiltIn("GPROP", "", 2, _bi_Gprop)
	workspace.registerBuiltIn("PLIST", "", 1, _bi_Plist)
	workspace.registerBuiltIn("PPROP", "", 3, _bi_Pprop)
	workspace.registerBuiltIn("PPS", "", 0, _bi_Pps)
	workspace.registerBuiltIn("REMPROP", "", 2, _bi_Remprop)
}
