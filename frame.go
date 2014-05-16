package main

import (
	"strings"
)

type Variable struct {
	name   string
	value  Node
	buried bool
}

type Frame interface {
	workspace() *Workspace
	parentFrame() Frame
	depth() int
	caller() *WordNode
	eval(parameters []Node) (Node, error)
	setTestValue(node Node)
	getTestValue() Node
	createLocal(name string)
	setVariable(name string, value Node)
	getVariable(name string) Node
}

type Procedure interface {
	parameterCount() int
	createFrame(parentFrame Frame, caller *WordNode) Frame
}

type evaluator func(Frame, []Node) (Node, error)

type BuiltInProcedure struct {
	name       string
	paramCount int
	realProc   evaluator
}

func (this *BuiltInProcedure) parameterCount() int {
	return this.paramCount
}

func (this *BuiltInProcedure) createFrame(parentFrame Frame, caller *WordNode) Frame {
	return &BuiltInFrame{parentFrame.workspace(), parentFrame, parentFrame.depth() + 1, caller, this.realProc, make(map[string]*Variable), this.name}
}

type BuiltInFrame struct {
	ws         *Workspace
	parent     Frame
	d          int
	callerNode *WordNode
	realProc   evaluator
	vars       map[string]*Variable
	name       string
}

func (this *BuiltInFrame) parentFrame() Frame {
	return this.parent
}

func (this *BuiltInFrame) workspace() *Workspace {
	return this.ws
}

func (this *BuiltInFrame) depth() int {
	return this.d
}

func (this *BuiltInFrame) caller() *WordNode {
	return this.callerNode
}

func (this *BuiltInFrame) eval(parameters []Node) (Node, error) {
	return this.realProc(this, parameters)
}

func (this *BuiltInFrame) setTestValue(node Node) {
}

func (this *BuiltInFrame) getTestValue() Node {
	return nil
}

func (this *BuiltInFrame) createLocal(name string) {
	this.vars[strings.ToUpper(name)] = &Variable{name, nil, false}
}

func (this *BuiltInFrame) setVariable(name string, value Node) {

	uname := strings.ToUpper(name)
	v, exists := this.vars[uname]
	if exists {
		v.value = value
	} else if this.parent == nil {
		this.vars[uname] = &Variable{name, value, false}
	} else {
		this.parent.setVariable(name, value)
	}
}

func (this *BuiltInFrame) getVariable(name string) Node {

	v, exists := this.vars[strings.ToUpper(name)]
	if exists {
		return v.value
	} else if this.parent == nil {
		return nil
	} else {
		return this.parent.getVariable(name)
	}
}

type RootFrame struct {
	ws      *Workspace
	node    Node
	testVal Node
	vars    map[string]*Variable
}

func (this *RootFrame) workspace() *Workspace {
	return this.ws
}

func (this *RootFrame) depth() int { return 0 }

func (this *RootFrame) parentFrame() Frame {
	return nil
}

func (this *RootFrame) caller() *WordNode {
	return nil
}

func (this *RootFrame) eval(parameters []Node) (Node, error) {

	ln := newListNode(-1, -1, this.node)
	return evalInstructionList(this, ln, false)
}

func (this *RootFrame) setTestValue(node Node) {
	this.testVal = node
}

func (this *RootFrame) getTestValue() Node {
	return this.testVal
}

func (this *RootFrame) createLocal(name string) {
	this.vars[strings.ToUpper(name)] = &Variable{name, nil, false}
}

func (this *RootFrame) setVariable(name string, value Node) {

	uname := strings.ToUpper(name)
	v, exists := this.vars[uname]
	if exists {
		v.value = value
	} else {
		this.vars[uname] = &Variable{name, value, false}
	}
}

func (this *RootFrame) getVariable(name string) Node {

	v, exists := this.vars[strings.ToUpper(name)]
	if exists {
		return v.value
	} else {
		return nil
	}
}

type InterpretedFrame struct {
	ws         *Workspace
	parent     Frame
	d          int
	callerNode *WordNode
	procedure  *InterpretedProcedure
	returnVal  Node
	testVal    Node
	vars       map[string]*Variable
	exit       bool
}

func (this *InterpretedFrame) workspace() *Workspace {
	return this.ws
}

func (this *InterpretedFrame) parentFrame() Frame {
	return this.parent
}

func (this *InterpretedFrame) caller() *WordNode {
	return this.callerNode
}

func (this *InterpretedFrame) depth() int {
	return this.d
}

func (this *InterpretedFrame) eval(parameters []Node) (Node, error) {

	for px := 0; px < len(parameters); px++ {
		this.createLocal(this.procedure.parameters[px])
		this.setVariable(this.procedure.parameters[px], parameters[px])
	}

	var err error
	for n := this.procedure.firstNode; n != nil; {
		_, n, err = callProcedure(this, n, true)
		if err != nil {
			return nil, err
		}
		if this.exit {
			break
		}
	}

	return this.returnVal, nil
}

func (this *InterpretedFrame) setReturnValue(returnVal Node) {
	this.returnVal = returnVal
	this.exit = true
}

func (this *InterpretedFrame) stop() {
	this.exit = true
}

func (this *InterpretedFrame) setTestValue(node Node) {
	this.testVal = node
}

func (this *InterpretedFrame) getTestValue() Node {
	return this.testVal
}

func (this *InterpretedFrame) createLocal(name string) {
	this.vars[strings.ToUpper(name)] = &Variable{name, nil, false}
}

func (this *InterpretedFrame) setVariable(name string, value Node) {

	uname := strings.ToUpper(name)
	v, exists := this.vars[uname]
	if exists {
		v.value = value
	} else if this.parent == nil {
		this.vars[uname] = &Variable{name, value, false}
	} else {
		this.parent.setVariable(name, value)
	}
}

func (this *InterpretedFrame) getVariable(name string) Node {

	v, exists := this.vars[strings.ToUpper(name)]
	if exists {
		return v.value
	} else if this.parent == nil {
		return nil
	} else {
		return this.parent.getVariable(name)
	}
}

func callProcedure(frame Frame, node Node, withInfix bool) (Node, Node, error) {

	wn := node.(*WordNode)
	if wn == nil {
		return nil, nil, errorWordExpected(node)
	}

	var proc Procedure
	var parameters []Node
	if wn.value[0] == ':' {
		proc = frame.workspace().findProcedure(keywordThing)
		c, r := wn.position()
		parameters = []Node{newWordNode(c, r, string(wn.value[1:]), true)}
		node = node.next()
	} else {
		if wn.isLiteral {
			return nil, nil, errorProcedureExpected(node)
		}
		procName := strings.ToUpper(wn.value)
		proc = frame.workspace().findProcedure(procName)

		if proc == nil {
			return nil, nil, errorProcedureNotFound(node, wn.value)
		}
		var err error
		if proc.parameterCount() > 0 {
			parameters, node, err = fetchParameters(frame, wn, node.next(), proc.parameterCount(), withInfix)
			if err != nil {
				return nil, nil, err
			}
		} else {
			parameters = make([]Node, 0, 0)
			node = node.next()
		}
	}

	subFrame := proc.createFrame(frame, wn)

	rv, err := subFrame.eval(parameters)
	if err != nil {
		return nil, nil, err
	}

	return rv, node, nil
}

func evaluateList(frame Frame, list *ListNode) (Node, error) {

	procFrame, _ := findInterpretedFrame(frame)

	var fn Node = nil
	var cn Node = nil
	var tn Node = nil
	var err error = nil
	for n := list.firstChild; n != nil; {
		tn, n, err = evaluateNode(frame, n, true)
		if err != nil {
			return nil, err
		}

		if fn == nil {
			fn = tn
		} else {
			cn.addNode(tn)
		}
		cn = tn

		if procFrame != nil && procFrame.exit {
			break
		}
	}

	return newListNode(list.line, list.col, fn), nil
}

func evalInstructionList(frame Frame, node Node, canReturn bool) (Node, error) {

	switch ln := node.(type) {
	case *WordNode:
		return nil, errorListExpected(node)
	case *ListNode:
		procFrame, _ := findInterpretedFrame(frame)
		var v Node
		var err error
		for n := ln.firstChild; n != nil; {
			v, n, err = evaluateNode(frame, n, true)
			if err != nil {
				return nil, err
			}
			if v != nil {
				if canReturn {
					return v, nil
				}
				return nil, errorReturnValueUnused(v)
			}
			if procFrame != nil && procFrame.exit {
				return nil, nil
			}
		}
	}

	return nil, nil
}

var infixOps []string = []string{"+", "-", "*", "/", "(", ")", "=", "<>", "<", ">", "<=", ">=", "OR", "AND"}
var infixProc []string = []string{"SUM", "DIFFERENCE", "PRODUCT", "QUOTIENT", "", "", "EQUALP", "NOTEQUALP", "LESSP", "GREATERP", "LESSEQUALP", "GREATEREQUALP", "EITHER", "BOTH"}
var infixPrec []int = []int{1, 1, 2, 2, 3, 4, 4, 5, 5, 5, 5, 5, 5, 6, 6}

func getInfixOp(nodeVal string) int {
	if len(nodeVal) > 3 {
		return -1
	}
	uv := strings.ToUpper(nodeVal)
	for ix := 0; ix < len(infixOps); ix++ {
		if infixOps[ix] == uv {
			return ix
		}
	}

	return -1
}

func evaluateExpression(frame Frame, n Node) (Node, Node, error) {

	nl := make([]Node, 0, 2)
	ops := make([]string, 0, 2)
	expectOp := false
	exit := false
	var prevIx int = -2
	var err error = nil
	braceCount := 0
	for !exit && n != nil {

		switch nn := n.(type) {
		case *WordNode:
			ix := getInfixOp(nn.value)
			if ix >= 0 {
				if ix == 4 {
					if prevIx == -1 || prevIx == 5 {
						exit = true
						break
					}
					braceCount++
				}
				if ix == 5 {
					braceCount--
					if braceCount < 0 {
						exit = true
						break
					}
				}
				if nn.value == "-" && (prevIx == -2 || (prevIx >= 0 && prevIx != 5)) && nn.next() != nil {
					// Looks like a unary minus
					var vn Node
					vn, n, err = evaluateNode(frame, nn.next(), true)
					if err != nil {
						return nil, nil, err
					}
					v, err := evalToNumber(vn)
					if err != nil {
						return nil, nil, err
					}

					nwn := createNumericNode(0.0 - v).(*WordNode)
					nwn.isInfix = true
					nl = append(nl, nwn)
					expectOp = true
				} else {
					prec := infixPrec[ix]
					if len(ops) > 0 {
						pop := ops[len(ops)-1]
						popIx := getInfixOp(pop)
						for (nn.value == ")" && pop != "(") || prec <= infixPrec[popIx] {
							l, c := nn.position()
							procName := infixProc[popIx]
							if procName != "" {
								nwn := newWordNode(l, c, procName, false).(*WordNode)
								nwn.isInfix = true
								nl = append(nl, nwn)
							}
							ops = ops[0 : len(ops)-1]
							if len(ops) == 0 {
								break
							}
							pop = ops[len(ops)-1]
							popIx = getInfixOp(pop)
						}
					}
					ops = append(ops, nn.value)
					n = n.next()
					expectOp = false
				}
			} else {
				if expectOp {
					exit = true
				} else {

					if nn.isLiteral {
						nl = append(nl, nn.clone())
						n = n.next()
						expectOp = true
					} else {
						var p Node
						var err error
						p, n, err = callProcedure(frame, nn, true)
						if err != nil {
							return nil, nil, err
						}
						if p != nil {
							nl = append(nl, p)
							expectOp = true
						}
					}
				}
			}
			prevIx = ix
		case *ListNode:
			exit = true
			break
		}
	}

	for len(ops) > 0 {
		pop := ops[len(ops)-1]
		popIx := getInfixOp(pop)

		procName := infixProc[popIx]
		if procName != "" {
			nwn := newWordNode(-1, -1, procName, false).(*WordNode)
			nwn.isInfix = true
			nl = append(nl, nwn)
		}
		ops = ops[0 : len(ops)-1]
	}

	if len(nl) == 1 {
		nl[0].addNode(n)
		return nl[0], n, nil
	} else if len(nl) > 1 {

		ix := len(nl) - 1
		fn := nl[ix]
		nn := fn
		ix--
		for ix >= 0 {
			nn.addNode(nl[ix])
			nn = nn.next()
			ix--
		}

		res, _, err := callProcedure(frame, fn, false)
		if res != nil {
			res.addNode(n)
		}
		return res, n, err
	}
	return nil, n, nil
}

func evaluateNode(frame Frame, node Node, withInfix bool) (Node, Node, error) {

	switch nn := node.(type) {
	case *WordNode:
		if withInfix {
			return evaluateExpression(frame, node)
		} else {
			if nn.isLiteral {
				return nn, nn.next(), nil
			} else {
				p, node, err := callProcedure(frame, node, withInfix)
				if err != nil {
					return nil, nil, err
				}
				return p, node, nil
			}
		}
	case *ListNode:
		return nn, node.next(), nil
	}

	return nil, nil, nil
}

func fetchParameters(frame Frame, caller *WordNode, firstNode Node, paramCount int, withInfix bool) ([]Node, Node, error) {
	params := make([]Node, 0, paramCount)

	n := firstNode
	var p Node
	var err error
	for ix := 0; ix < paramCount; ix++ {
		if n == nil {
			return nil, nil, errorNotEnoughParameters(caller, firstNode)
		}

		switch nn := n.(type) {
		case *WordNode:
			p, n, err = evaluateNode(frame, nn, withInfix)
			if err != nil {
				return nil, nil, err
			}
			params = append(params, p)

		case *ListNode:
			params = append(params, nn)
			n = nn.next()
		}
	}
	return params, n, nil
}

type InterpretedProcedure struct {
	name       string
	parameters []string
	firstNode  Node
	source     string
	buried     bool
}

func (this *InterpretedProcedure) createFrame(parentFrame Frame, caller *WordNode) Frame {

	parentFrame.workspace().trace(parentFrame.depth(), this.name)
	return &InterpretedFrame{parentFrame.workspace(), parentFrame, parentFrame.depth() + 1, caller, this, nil, nil, make(map[string]*Variable), false}
}

func (this *InterpretedProcedure) parameterCount() int {
	return len(this.parameters)
}

func readInterpretedProcedure(node Node) (*InterpretedProcedure, Node, error) {

	if !isWordNodeWithValue(node, keywordTo) {
		return nil, nil, errorKeywordExpected(node, keywordTo)
	}

	procName, n, err := readWordValue(node.next())
	if err != nil {
		return nil, nil, err
	}

	params := make([]string, 0, 2)
	for ; n != nil; n = n.next() {
		wn := n.(*WordNode)
		if wn == nil {
			break
		}
		if wn.value[0] != ':' {
			break
		}
		params = append(params, wn.value[1:])
	}

	firstNode := n
	isComplete := false
	var pn Node = nil
	for ; n != nil; n = n.next() {

		if isWordNodeWithValue(n, keywordEnd) {
			if pn != nil {
				pn.addNode(nil)
			}
			isComplete = true
			break
		}

		pn = n
	}

	if !isComplete {
		return nil, nil, errorKeywordExpected(nil, keywordEnd)
	}

	return &InterpretedProcedure{strings.ToUpper(procName), params, firstNode, "", false}, n.next(), nil
}

func findInterpretedFrame(frame Frame) (*InterpretedFrame, error) {

	orig := frame

	for {
		switch f := frame.(type) {
		case *InterpretedFrame:
			return f, nil
		case *RootFrame:
			c := orig.caller()
			if c == nil {
				return nil, nil
			}
			return nil, errorNoInterpretedFrame(c)
		case *BuiltInFrame:
			frame = frame.parentFrame()

		}
	}

	return nil, errorNoInterpretedFrame(orig.caller())
}
