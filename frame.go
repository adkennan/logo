package main

import (
	"strings"
)

type Variable struct {
	name   string
	value  Node
	buried bool
	props  map[string]Node
}

func newVar(name string) *Variable {
	return &Variable{name, nil, false, nil}
}

func newVarVal(name string, value Node) *Variable {
	return &Variable{name, value, false, nil}
}

func (this *Variable) hasProps() bool { return this.props != nil }

func (this *Variable) setProp(name string, value Node) {
	if this.props == nil {
		this.props = make(map[string]Node)
	}
	this.props[name] = value
}

func (this *Variable) getProp(name string) Node {
	if this.props == nil {
		return nil
	}

	val, ok := this.props[name]
	if !ok {
		return nil
	}

	return val
}

func (this *Variable) clearProp(name string) {
	if this.props == nil {
		return
	}

	delete(this.props, name)
}

func (this *Variable) clearProps() {
	if this.props == nil {
		return
	}
	this.props = nil
}

type VarList struct {
	vars map[string]*Variable
}

func newVarList() *VarList {
	return &VarList{nil}
}

func (this *VarList) createLocal(name string) {
	if this.vars == nil {
		this.vars = make(map[string]*Variable)
	}
	this.vars[strings.ToUpper(name)] = newVar(name)
}

func (this *VarList) setVariable(frame Frame, name string, value Node) {

	uname := strings.ToUpper(name)
	v := this.getVariableInner(frame, name, uname, true)
	v.value = value
}

func (this *VarList) getVariable(frame Frame, name string) Node {

	uname := strings.ToUpper(name)
	v := this.getVariableInner(frame, name, uname, false)
	if v == nil {
		return nil
	}
	return v.value
}

func (this *VarList) getVariableInner(frame Frame, name, uname string, canCreate bool) *Variable {

	if this.vars != nil {
		v, exists := this.vars[uname]
		if exists {
			return v
		}
	}
	pf := frame.parentFrame()
	if pf == nil {
		if canCreate {
			if this.vars == nil {
				this.vars = make(map[string]*Variable)
			}
			v := newVar(name)
			this.vars[uname] = v
			return v
		}
		return nil
	}
	return pf.getVars().getVariableInner(pf, name, uname, canCreate)
}

func (this *VarList) hasProps(frame Frame, name string) bool {

	uname := strings.ToUpper(name)
	v := this.getVariableInner(frame, name, uname, false)
	if v == nil {
		return false
	}
	return v.hasProps()
}

func (this *VarList) setProp(frame Frame, varName, propName string, value Node) {

	uname := strings.ToUpper(varName)
	v := this.getVariableInner(frame, varName, uname, true)
	v.setProp(propName, value)
}

func (this *VarList) getProp(frame Frame, varName, propName string) Node {
	uname := strings.ToUpper(varName)
	v := this.getVariableInner(frame, varName, uname, false)
	if v == nil {
		return nil
	}
	return v.getProp(propName)
}

func (this *VarList) clearProp(frame Frame, varName, propName string) {
	uname := strings.ToUpper(varName)
	v := this.getVariableInner(frame, varName, uname, false)
	if v == nil {
		return
	}
	v.clearProp(propName)
}

func (this *VarList) clearProps() {
	if this.vars == nil {
		return
	}

	for _, v := range this.vars {
		if v.hasProps() {
			v.clearProps()
		}
	}
}

func (this *VarList) setBuried(frame Frame, name string, buried bool) {

	uname := strings.ToUpper(name)
	v := this.getVariableInner(frame, name, uname, false)
	if v == nil {
		return
	}
	v.buried = buried
}

func (this *VarList) isBuried(frame Frame, name string) bool {
	uname := strings.ToUpper(name)
	v := this.getVariableInner(frame, name, uname, false)
	if v == nil {
		return false
	}
	return v.buried
}

func (this *VarList) setAllBuried(frame Frame, buried bool) {
	if this.vars == nil {
		return
	}

	for _, v := range this.vars {
		v.buried = buried
	}
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
	getVars() *VarList
	isStepped() bool
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
	return &BuiltInFrame{parentFrame.workspace(), parentFrame, parentFrame.depth() + 1, caller, this.realProc, newVarList(), this.name}
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
	vars       *VarList
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

func (this *BuiltInFrame) getVars() *VarList {
	return this.vars
}

func (this *BuiltInFrame) isStepped() bool { return false }

type RootFrame struct {
	ws      *Workspace
	node    Node
	testVal Node
	vars    *VarList
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

func (this *RootFrame) getVars() *VarList {
	return this.vars
}

func (this *RootFrame) isStepped() bool { return false }

type InterpretedFrame struct {
	ws         *Workspace
	parent     Frame
	d          int
	callerNode *WordNode
	procedure  *InterpretedProcedure
	returnVal  Node
	testVal    Node
	vars       *VarList
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
		this.vars.createLocal(this.procedure.parameters[px])
		this.vars.setVariable(this, this.procedure.parameters[px], parameters[px])
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

func (this *InterpretedFrame) getVars() *VarList {
	return this.vars
}

func (this *InterpretedFrame) isStepped() bool {
	return this.procedure.step
}

func (this *InterpretedFrame) step(currentNode Node) *CallResult {
	cl, _ := currentNode.position()
	e := EnumerateWords(this.procedure.firstNode)
	n := e.nextWord()
	for n != nil {
		l, _ := n.position()
		if l == cl {
			printLine(this.ws, n, currentNode)
			c, err := this.ws.files.reader.ReadChar()
			if err != nil {
				return errorResult(err)
			}
			if c == K_ESCAPE {
				return stopResult()
			}
			return nil
		}
		n = e.nextWord()
	}

	println("duh")
	return nil
}

type InterpretedProcedure struct {
	name       string
	parameters []string
	firstNode  Node
	source     string
	buried     bool
	step       bool
}

func (this *InterpretedProcedure) createFrame(parentFrame Frame, caller *WordNode) Frame {

	parentFrame.workspace().trace(parentFrame.depth(), this.name)
	return &InterpretedFrame{parentFrame.workspace(), parentFrame, parentFrame.depth() + 1, caller, this, nil, nil, newVarList(), false, false}
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

	return &InterpretedProcedure{strings.ToUpper(procName), params, firstNode, "", false, false}, n.next(), nil
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
