package logo

import (
	"bufio"
	"fmt"
	"math"
	"math/rand"
	"os"
)

var keywordTo string = "TO"
var keywordEnd string = "END"
var keywordTrue string = "TRUE"
var keywordFalse string = "FALSE"
var keywordThing string = "THING"

var trueNode Node = newWordNode(-1, -1, keywordTrue, true)
var falseNode Node = newWordNode(-1, -1, keywordFalse, true)
var randomMax Node = newWordNode(-1, -1, "10", true)

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
	fmt.Print("\n")
	return nil, nil
}

func _bi_FPrint(frame Frame, parameters []Node) (Node, error) {

	printNode(parameters[0], true)
	fmt.Print("\n")
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

	if nodesEqual(x, y) {
		return trueNode, nil
	}
	return falseNode, nil
}

func _bi_NotEqualp(frame Frame, parameters []Node) (Node, error) {

	x := parameters[0]
	y := parameters[1]

	if nodesEqual(x, y) {
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

	if nodesEqual(n, trueNode) {
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

	if nodesEqual(n, falseNode) {
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

	return scope
}
