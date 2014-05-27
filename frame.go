package main

import (
	"strings"
)

type Variable struct {
	name   string
	value  Node
	buried bool
}

type CallResult struct {
	returnValue Node
	err         error
	stopped     bool
}

func (this *CallResult) hasError() bool { return this.err != nil }

func (this *CallResult) hasReturnVal() bool { return this.returnValue != nil }

func (this *CallResult) shouldStop() bool { return this.err != nil || this.stopped }

func returnResult(node Node) *CallResult {
	return &CallResult{node, nil, false}
}

func errorResult(err error) *CallResult {
	return &CallResult{nil, err, false}
}

func stopResult() *CallResult {
	return &CallResult{nil, nil, true}
}

type Frame interface {
	workspace() *Workspace
	parentFrame() Frame
	depth() int
	caller() *WordNode
	eval(parameters []Node) *CallResult
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

type evaluator func(Frame, []Node) *CallResult

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

func (this *BuiltInFrame) eval(parameters []Node) *CallResult {

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

func (this *RootFrame) eval(parameters []Node) *CallResult {

	ln := newListNode(-1, -1, this.node)
	return evaluateList(this, ln)
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
	stopped    bool
	aborted    bool
}

func (this *InterpretedFrame) abort() {
	this.aborted = true
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

func (this *InterpretedFrame) eval(parameters []Node) *CallResult {

	for px := 0; px < len(parameters); px++ {
		this.createLocal(this.procedure.parameters[px])
		this.setVariable(this.procedure.parameters[px], parameters[px])
	}

	var rv *CallResult
	for n := this.procedure.firstNode; n != nil; {
		rv, n = callProcedure(this, n, true)
		if this.aborted {
			return errorResult(errorUserStopped(this.callerNode))
		}
		if rv != nil {
			if rv.hasError() {
				return rv
			}
		}

		if this.stopped {
			break
		}
	}

	if this.returnVal == nil {
		return nil
	}
	return returnResult(this.returnVal)
}

func (this *InterpretedFrame) setReturnValue(returnVal Node) {
	this.returnVal = returnVal
	this.stopped = true
}

func (this *InterpretedFrame) stop() {
	this.stopped = true
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

	var n Node
	v, exists := this.vars[strings.ToUpper(name)]
	if exists {
		n = v.value
	} else if this.parent != nil {
		n = this.parent.getVariable(name)
	}

	return n
}

func isProcedure(wn *WordNode, ws *Workspace) bool {
	if wn.value[0] == ':' {
		return true
	}
	procName := strings.ToUpper(wn.value)
	return ws.findProcedure(procName) != nil
}

func callProcedure(frame Frame, node Node, withInfix bool) (*CallResult, Node) {

	wn := node.(*WordNode)
	if wn == nil {
		return errorResult(errorWordExpected(node)), nil
	}

	var proc Procedure
	var parameters []Node
	var procName string
	if wn.value[0] == ':' {
		proc = frame.workspace().findProcedure(keywordThing)
		procName = keywordThing
		c, r := wn.position()
		parameters = []Node{newWordNode(c, r, string(wn.value[1:]), true)}
		node = node.next()
	} else {
		if wn.isLiteral {
			return errorResult(errorProcedureExpected(node)), nil
		}
		procName = strings.ToUpper(wn.value)
		proc = frame.workspace().findProcedure(procName)

		if proc == nil {
			return errorResult(errorProcedureNotFound(node, wn.value)), nil
		}
		var err error
		if proc.parameterCount() > 0 {
			parameters, node, err = fetchParameters(frame, wn, node.next(), proc.parameterCount(), withInfix)
			if err != nil {
				return errorResult(err), nil
			}
		} else {
			parameters = make([]Node, 0, 0)
			node = node.next()
		}
	}

	subFrame := proc.createFrame(frame, wn)
	frame.workspace().currentFrame = subFrame
	defer func() { frame.workspace().currentFrame = frame }()

	print("Calling ", procName, "(")
	for _, p := range parameters {
		print(p.String(), " ")
	}
	print(") = ")

	rv := subFrame.eval(parameters)

	if rv != nil {
		if rv.hasError() {
			println("ERROR: ", rv.err.Error())
			return rv, nil
		} else if rv.returnValue != nil {
			println("{ ", rv.returnValue.String(), " }")
		} else {
			println("<NIL>")
		}
	} else {
		println("<NIL>")
	}
	return rv, node
}

func evaluateList(frame Frame, list *ListNode) *CallResult {

	var fn Node = nil
	var cn Node = nil
	var rv *CallResult
	for n := list.firstChild; n != nil; {
		rv, n = evaluateNode(frame, n, true)
		if rv != nil {
			if rv.hasError() {
				return rv
			}
			if rv.stopped {
				break
			}

			if fn == nil {
				fn = rv.returnValue
			} else {
				cn.addNode(rv.returnValue)
			}
			cn = rv.returnValue
		}
	}

	return returnResult(newListNode(list.line, list.col, fn))
}

func evalInstructionList(frame Frame, node Node, canReturn bool) *CallResult {

	switch ln := node.(type) {
	case *WordNode:
		return errorResult(errorListExpected(node))
	case *ListNode:
		var r *CallResult
		for n := ln.firstChild; n != nil; {
			r, n = evaluateNode(frame, n, true)
			if r != nil {
				if r.shouldStop() {
					return r
				}
				if r.hasReturnVal() {
					if canReturn {
						return r
					}
					return errorResult(errorReturnValueUnused(r.returnValue))
				}
			}
		}
	}

	return nil
}

var infixOps []string = []string{"+", "-", "*", "/", "(", ")", "=", "<>", "<", ">", "<=", ">=", "OR", "AND"}
var infixProc []string = []string{"SUM", "DIFFERENCE", "PRODUCT", "QUOTIENT", "", "", "EQUALP", "NOTEQUALP", "LESSP", "GREATERP", "LESSEQUALP", "GREATEREQUALP", "EITHER", "BOTH"}
var infixPrec []int = []int{1, 1, 2, 2, 3, 4, 4, 0, 0, 0, 0, 0, 0, 0, 0}

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

func evaluateExpression(frame Frame, n Node) (*CallResult, Node) {

	nl := make([]Node, 0, 2)
	ops := make([]string, 0, 2)
	expectOp := false
	exit := false
	var prevIx int = -2
	braceCount := 0
	for !exit && n != nil {

		switch nn := n.(type) {
		case *WordNode:
			ix := getInfixOp(nn.value)
			if ix >= 0 {
				if ix == 4 {
					if prevIx == -1 || prevIx == 5 {
						println("exit 1")
						exit = true
						break
					}
					braceCount++
				}
				if ix == 5 {
					braceCount--
					if braceCount < 0 {
						println("exit 2")
						exit = true
						break
					}
				}
				if nn.value == "-" && (prevIx == -2 || (prevIx >= 0 && prevIx != 5)) && nn.next() != nil {
					// Looks like a unary minus
					var rv *CallResult
					rv, n = evaluateNode(frame, nn.next(), true)
					if rv.shouldStop() {
						return rv, nil
					}
					v, err := evalToNumber(rv.returnValue)
					if err != nil {
						return errorResult(err), nil
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
								nwn := newWordNode(l, c, procName, false)
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
				} else if !nn.isLiteral {
					var rv *CallResult
					rv, n = callProcedure(frame, nn, true)
					if rv != nil {
						if rv.shouldStop() {
							return rv, nil
						}
						if rv.returnValue != nil {
							nl = append(nl, rv.returnValue.clone())
							expectOp = true
						}
					}
				} else {
					nl = append(nl, nn.clone())
					n = n.next()
					expectOp = true
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
			nwn := newWordNode(-1, -1, procName, false)
			nwn.isInfix = true
			nl = append(nl, nwn)
		}
		ops = ops[0 : len(ops)-1]
	}

	var rv *CallResult
	if len(nl) == 1 {
		println("a:", nl[0].String())
		rv, _ = evaluateNode(frame, nl[0], false)
	} else if len(nl) > 1 {

		ix := len(nl) - 1
		fn := nl[ix]
		nn := fn
		ix--
		print("b:", fn.String(), " ")
		for ix >= 0 {
			print(nl[ix].String(), " ")
			nn.addNode(nl[ix])
			nn = nn.next()
			ix--
		}
		println()

		rv, _ = evaluateNode(frame, fn, false)
	}
	if rv != nil {
		if rv.shouldStop() {
			return rv, nil
		}
		if rv.returnValue != nil {
			rv.returnValue.addNode(n)
		}
	}
	return rv, n
}

func evaluateNode(frame Frame, node Node, withInfix bool) (*CallResult, Node) {

	switch nn := node.(type) {
	case *WordNode:
		if withInfix {
			return evaluateExpression(frame, node)
		} else {
			if !nn.isLiteral {
				rv, node := callProcedure(frame, node, withInfix)
				if rv != nil && rv.shouldStop() {
					return rv, nil
				}
				return rv, node
			} else {
				return returnResult(nn), nn.next()
			}
		}
	case *ListNode:
		return returnResult(nn), node.next()
	}

	return nil, nil
}

func fetchParameters(frame Frame, caller *WordNode, firstNode Node, paramCount int, withInfix bool) ([]Node, Node, error) {
	params := make([]Node, 0, paramCount)

	n := firstNode
	var rv *CallResult
	for ix := 0; ix < paramCount; ix++ {
		if n == nil {
			return nil, nil, errorNotEnoughParameters(caller, firstNode)
		}

		switch nn := n.(type) {
		case *WordNode:
			rv, n = evaluateNode(frame, nn, withInfix)
			if rv != nil {
				if rv.hasError() {
					return nil, nil, rv.err
				}
				params = append(params, rv.returnValue)
			}

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
	return &InterpretedFrame{parentFrame.workspace(), parentFrame, parentFrame.depth() + 1, caller, this, nil, nil, make(map[string]*Variable), false, false}
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
