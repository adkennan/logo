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
	allowVarParameters() bool
	createFrame(parentFrame Frame, caller *WordNode) Frame
}

type evaluator func(Frame, []Node) *CallResult

type BuiltInProcedure struct {
	name           string
	paramCount     int
	allowVarParams bool
	realProc       evaluator
}

func (this *BuiltInProcedure) parameterCount() int {
	return this.paramCount
}

func (this *BuiltInProcedure) createFrame(parentFrame Frame, caller *WordNode) Frame {
	return &BuiltInFrame{parentFrame.workspace(), parentFrame, parentFrame.depth() + 1, caller, this.realProc, make(map[string]*Variable), this.name}
}

func (this *BuiltInProcedure) allowVarParameters() bool {
	return this.allowVarParams
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

	return evalNodeStream(this, this.node, false)
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
	this.stopped = true
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

	if this.procedure.firstNode != nil {
		rv := evalNodeStream(this, this.procedure.firstNode, false)
		if rv != nil && rv.hasError() {
			return rv
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

func (this *InterpretedProcedure) allowVarParameters() bool {
	return false
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
