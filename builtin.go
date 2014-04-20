package logo

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
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

var OutputHandler func(string) error
var TraceHandler func(int, string)

var ExitHandler func()

var traceEnabled bool

func Print(text string) error {
	if OutputHandler != nil {
		return OutputHandler(text)
	}

	_, err := fmt.Print(text)
	return err
}

func _bi_Repeat(frame Frame, parameters []Node) (Node, error) {

	n, err := evalToNumber(parameters[0])
	if err != nil {
		return nil, err
	}
	nn := int64(n)

	for ix := int64(0); ix < nn; ix++ {
		evalInstructionList(frame, parameters[1])
	}

	return nil, nil
}

func _bi_Print(frame Frame, parameters []Node) (Node, error) {

	printNode(parameters[0], false)
	Print("\n")
	return nil, nil
}

func _bi_FPrint(frame Frame, parameters []Node) (Node, error) {

	printNode(parameters[0], true)
	Print("\n")
	return nil, nil
}

func _bi_Type(frame Frame, parameters []Node) (Node, error) {

	printNode(parameters[0], false)
	return nil, nil
}

func _bi_FType(frame Frame, parameters []Node) (Node, error) {

	printNode(parameters[0], true)
	return nil, nil
}

func _bi_If(frame Frame, parameters []Node) (Node, error) {

	r, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}

	if r {
		return nil, evalInstructionList(frame, parameters[1])
	}

	return nil, nil
}

func _bi_ReadList(frame Frame, parameters []Node) (Node, error) {

	bio := bufio.NewReader(os.Stdin)
	line, _, err := bio.ReadLine()
	if err != nil {
		return nil, err
	}
	return ParseString("[ " + string(line) + " ]")
}

func _bi_Request(frame Frame, parameters []Node) (Node, error) {

	bio := bufio.NewReader(os.Stdin)
	fmt.Print("? ")
	line, _, err := bio.ReadLine()
	if err != nil {
		return nil, err
	}
	return ParseString("[ " + string(line) + " ]")
}

func _bi_IfElse(frame Frame, parameters []Node) (Node, error) {

	r, err := evalToBoolean(parameters[0])
	if err != nil {
		return nil, err
	}

	if r {
		return nil, evalInstructionList(frame, parameters[1])
	} else {
		return nil, evalInstructionList(frame, parameters[2])
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

	x, y, err := evalNumericParams(parameters[0], parameters[1])
	if err != nil {
		return nil, err
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

	if ex == nil && ey == nil && nx > ny {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_Lessp(frame Frame, parameters []Node) (Node, error) {

	nx, ex := evalToNumber(parameters[0])
	ny, ey := evalToNumber(parameters[1])

	if ex == nil && ey == nil && nx < ny {
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
		err := evalInstructionList(frame, parameters[0])
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func _bi_IfFalse(frame Frame, parameters []Node) (Node, error) {

	n := frame.parentFrame().getTestValue()
	if n == nil {
		return nil, errorNoCurrentTestValue(parameters[0])
	}

	if nodesEqual(n, falseNode, false) {
		err := evalInstructionList(frame, parameters[0])
		if err != nil {
			return nil, err
		}
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

	Print("Seeya!\n\n")

	if ExitHandler != nil {
		ExitHandler()
	} else {
		os.Exit(0)
	}
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

	traceEnabled = true

	return nil, nil
}

func _bi_Untrace(frame Frame, parameters []Node) (Node, error) {

	traceEnabled = false

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

	return nil, evalInstructionList(frame, parameters[0])
}

func CreateBuiltInScope() Scope {

	scope := &BuiltInScope{make(map[string]Procedure, 10), nil}

	scope.registerBuiltIn("PRINT", "PR", 1, _bi_Print)
	scope.registerBuiltIn("FPRINT", "FP", 1, _bi_FPrint)
	scope.registerBuiltIn("TYPE", "TY", 1, _bi_Type)
	scope.registerBuiltIn("FTYPE", "FTY", 1, _bi_FType)

	scope.registerBuiltIn("READLIST", "RL", 0, _bi_ReadList)
	scope.registerBuiltIn("REQUEST", "", 0, _bi_Request)

	scope.registerBuiltIn("REPEAT", "", 2, _bi_Repeat)
	scope.registerBuiltIn("IF", "", 2, _bi_If)
	scope.registerBuiltIn("IFELSE", "", 3, _bi_IfElse)

	scope.registerBuiltIn("SUM", "", 2, _bi_Sum)
	scope.registerBuiltIn("DIFFERENCE", "DIFF", 2, _bi_Difference)
	scope.registerBuiltIn("PRODUCT", "", 2, _bi_Product)
	scope.registerBuiltIn("QUOTIENT", "", 2, _bi_Quotient)
	scope.registerBuiltIn("REMAINDER", "MOD", 2, _bi_Remainder)
	scope.registerBuiltIn("MAXIMUM", "MAX", 2, _bi_Maximum)
	scope.registerBuiltIn("MINIMUM", "MIN", 2, _bi_Minimum)

	scope.registerBuiltIn("EQUALP", "EQUAL?", 2, _bi_Equalp)
	scope.registerBuiltIn("IS", "EQUAL?", 2, _bi_Is)
	scope.registerBuiltIn("NOTEQUALP", "NOTEQUAL?", 2, _bi_NotEqualp)
	scope.registerBuiltIn("GREATERP", "GREATER?", 2, _bi_Greaterp)
	scope.registerBuiltIn("LESSP", "LESS?", 2, _bi_Lessp)
	scope.registerBuiltIn("NUMBERP", "NUMBER?", 1, _bi_Numberp)
	scope.registerBuiltIn("ZEROP", "ZERO?", 1, _bi_Zerop)

	scope.registerBuiltIn("RANDOM", "", 0, _bi_Random)
	scope.registerBuiltIn("RND", "", 1, _bi_Rnd)
	scope.registerBuiltIn("SQRT", "", 1, _bi_Sqrt)
	scope.registerBuiltIn("POW", "", 2, _bi_Pow)
	scope.registerBuiltIn("SIN", "", 1, _bi_Sin)
	scope.registerBuiltIn("COS", "", 1, _bi_Cos)
	scope.registerBuiltIn("ARCTAN", "", 1, _bi_Arctan)

	scope.registerBuiltIn("TEST", "", 1, _bi_Test)
	scope.registerBuiltIn("IFTRUE", "", 1, _bi_IfTrue)
	scope.registerBuiltIn("IFFALSE", "", 1, _bi_IfFalse)

	scope.registerBuiltIn("MAKE", "", 2, _bi_Make)
	scope.registerBuiltIn("THING", "", 1, _bi_Thing)
	scope.registerBuiltIn("LOCAL", "", 1, _bi_Local)

	scope.registerBuiltIn("WORD", "", 2, _bi_Word)
	scope.registerBuiltIn("SENTENCE", "SE", 2, _bi_Sentence)
	scope.registerBuiltIn("LIST", "", 2, _bi_List)
	scope.registerBuiltIn("FPUT", "", 2, _bi_FPut)
	scope.registerBuiltIn("LPUT", "", 2, _bi_LPut)
	scope.registerBuiltIn("FIRST", "", 1, _bi_First)
	scope.registerBuiltIn("LAST", "", 1, _bi_Last)
	scope.registerBuiltIn("BUTFIRST", "", 1, _bi_ButFirst)
	scope.registerBuiltIn("BUTLAST", "", 1, _bi_ButLast)

	scope.registerBuiltIn("COUNT", "", 1, _bi_Count)
	scope.registerBuiltIn("EMPTYP", "", 1, _bi_Emptyp)
	scope.registerBuiltIn("WORDP", "", 1, _bi_Wordp)
	scope.registerBuiltIn("SENTENCEP", "", 1, _bi_Sentencep)
	scope.registerBuiltIn("MEMBERP", "", 2, _bi_Memberp)
	scope.registerBuiltIn("ITEM", "NTH", 2, _bi_Item)

	scope.registerBuiltIn("BOTH", "AND", 2, _bi_Both)
	scope.registerBuiltIn("EITHER", "OR", 2, _bi_Either)
	scope.registerBuiltIn("NOT", "", 1, _bi_Not)

	scope.registerBuiltIn("TRACE", "", 0, _bi_Trace)
	scope.registerBuiltIn("UNTRACE", "", 0, _bi_Untrace)
	scope.registerBuiltIn("GOODBYE", "BYE", 0, _bi_Goodbye)
	scope.registerBuiltIn("WAIT", "", 1, _bi_Wait)

	scope.registerBuiltIn("RUN", "", 1, _bi_Run)
	return scope
}
