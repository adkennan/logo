package logo

import (
	"strings"
)

type Variable struct {
	name  string
	value Node
}

type Frame interface {
	owningScope() Scope
	parentFrame() Frame
	eval(parameters []Node) (Node, error)
	setReturnValue(returnVal Node)
	setTestValue(node Node)
	getTestValue() Node
	createLocal(name string)
	setVariable(name string, value Node)
	getVariable(name string) Node
}

type Procedure interface {
	parameterCount() int
	createFrame(scope Scope, parentFrame Frame) Frame
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

func (this *BuiltInProcedure) createFrame(scope Scope, parentFrame Frame) Frame {
	return &BuiltInFrame{scope, parentFrame, this.realProc, nil, nil, make(map[string]*Variable)}
}

type BuiltInFrame struct {
	scope     Scope
	parent    Frame
	realProc  evaluator
	returnVal Node
	testVal   Node
	vars      map[string]*Variable
}

func (this *BuiltInFrame) owningScope() Scope {
	return this.scope
}

func (this *BuiltInFrame) parentFrame() Frame {
	return this.parent
}

func (this *BuiltInFrame) eval(parameters []Node) (Node, error) {
	return this.realProc(this, parameters)
}

func (this *BuiltInFrame) setReturnValue(returnVal Node) {
	this.returnVal = returnVal
}

func (this *BuiltInFrame) setTestValue(node Node) {
	this.testVal = node
}

func (this *BuiltInFrame) getTestValue() Node {
	return this.testVal
}

func (this *BuiltInFrame) createLocal(name string) {
	this.vars[strings.ToUpper(name)] = &Variable{name, nil}
}

func (this *BuiltInFrame) setVariable(name string, value Node) {

	uname := strings.ToUpper(name)
	v, exists := this.vars[uname]
	if exists {
		v.value = value
	} else if this.parent == nil {
		this.vars[uname] = &Variable{name, value}
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

type InteractiveFrame struct {
	scope   Scope
	parent  Frame
	node    Node
	testVal Node
	vars    map[string]*Variable
}

func (this *InteractiveFrame) owningScope() Scope {
	return this.scope
}

func (this InteractiveFrame) parentFrame() Frame {
	return this.parent
}

func (this *InteractiveFrame) eval(parameters []Node) (Node, error) {

	var err error
	var lastRetVal Node
	for n := this.node; n != nil; {
		lastRetVal, n, err = callProcedure(this, n)
		if err != nil {
			return nil, err
		}
	}

	return lastRetVal, nil
}

func (this *InteractiveFrame) setReturnValue(returnVal Node) {
}

func (this *InteractiveFrame) setTestValue(node Node) {
	this.testVal = node
}

func (this *InteractiveFrame) getTestValue() Node {
	return this.testVal
}

func (this *InteractiveFrame) createLocal(name string) {
	this.vars[strings.ToUpper(name)] = &Variable{name, nil}
}

func (this *InteractiveFrame) setVariable(name string, value Node) {

	uname := strings.ToUpper(name)
	v, exists := this.vars[uname]
	if exists {
		v.value = value
	} else if this.parent == nil {
		this.vars[uname] = &Variable{name, value}
	} else {
		this.parent.setVariable(name, value)
	}
}

func (this *InteractiveFrame) getVariable(name string) Node {

	v, exists := this.vars[strings.ToUpper(name)]
	if exists {
		return v.value
	} else if this.parent == nil {
		return nil
	} else {
		return this.parent.getVariable(name)
	}
}

var interactiveFrame *InteractiveFrame = &InteractiveFrame{nil, nil, nil, nil, make(map[string]*Variable)}

func Evaluate(scope Scope, source string) error {

	n, err := ParseString(source)
	if err != nil {
		return err
	}

	interactiveFrame.scope = scope
	interactiveFrame.node = n

	n, err = interactiveFrame.eval(make([]Node, 0, 0))
	if err != nil {
		return err
	}
	if n != nil {
		errorReturnValueUnused(interactiveFrame.node, n)
	}
	return nil
}

type InterpretedFrame struct {
	scope     Scope
	parent    Frame
	procedure *InterpretedProcedure
	returnVal Node
	testVal   Node
	vars      map[string]*Variable
}

func (this *InterpretedFrame) owningScope() Scope {
	return this.scope
}

func (this *InterpretedFrame) parentFrame() Frame {
	return this.parent
}

func (this *InterpretedFrame) eval(parameters []Node) (Node, error) {

	for px := 0; px < len(parameters); px++ {
		this.createLocal(this.procedure.parameters[px])
		this.setVariable(this.procedure.parameters[px], parameters[px])
	}

	var err error
	for n := this.procedure.firstNode; n != nil; {
		_, n, err = callProcedure(this, n)
		if err != nil {
			return nil, err
		}
	}

	return this.returnVal, nil
}

func (this *InterpretedFrame) setReturnValue(returnVal Node) {
	this.returnVal = returnVal
}

func (this *InterpretedFrame) setTestValue(node Node) {
	this.testVal = node
}

func (this *InterpretedFrame) getTestValue() Node {
	return this.testVal
}

func (this *InterpretedFrame) createLocal(name string) {
	this.vars[strings.ToUpper(name)] = &Variable{name, nil}
}

func (this *InterpretedFrame) setVariable(name string, value Node) {

	uname := strings.ToUpper(name)
	v, exists := this.vars[uname]
	if exists {
		v.value = value
	} else if this.parent == nil {
		this.vars[uname] = &Variable{name, value}
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

func callProcedure(frame Frame, node Node) (Node, Node, error) {

	wn := node.(*WordNode)
	if wn == nil {
		return nil, nil, errorWordExpected(node)
	}

	var proc Procedure
	var parameters []Node
	if wn.value[0] == ':' {
		proc = frame.owningScope().findProcedure(keywordThing)
		c, r := wn.position()
		parameters = []Node{newWordNode(c, r, string(wn.value[1:]), true)}
		node = node.next()
	} else {
		if wn.isLiteral {
			return nil, nil, errorProcedureExpected(node)
		}
		proc = frame.owningScope().findProcedure(strings.ToUpper(wn.value))

		if proc == nil {
			return nil, nil, errorProcedureNotFound(node, wn.value)
		}

		var err error
		if proc.parameterCount() > 0 {
			parameters, node, err = fetchParameters(frame, node.next(), proc.parameterCount())
			if err != nil {
				return nil, nil, err
			}
		} else {
			parameters = make([]Node, 0, 0)
			node = node.next()
		}
	}

	subFrame := proc.createFrame(frame.owningScope(), frame)

	rv, err := subFrame.eval(parameters)
	if err != nil {
		return nil, nil, err
	}

	return rv, node, nil
}

func evaluateList(frame Frame, list *ListNode) (Node, error) {

	var fn Node = nil
	var cn Node = nil
	var tn Node = nil
	var err error = nil
	for n := list.firstChild; n != nil; {
		tn, n, err = evaluateNode(frame, n)
		if err != nil {
			return nil, err
		}

		if fn == nil {
			fn = tn
		} else {
			cn.addNode(tn)
		}
		cn = tn
	}

	return newListNode(list.line, list.col, fn), nil
}

func evaluateNode(frame Frame, node Node) (Node, Node, error) {

	switch nn := node.(type) {
	case *WordNode:
		if nn.isLiteral {
			return nn, nn.next(), nil
		} else {
			p, node, err := callProcedure(frame, node)
			if err != nil {
				return nil, nil, err
			}
			return p, node, nil
		}
	case *ListNode:
		return nn, node.next(), nil
	}

	panic("Unknown node type!")
}

func fetchParameters(frame Frame, firstNode Node, paramCount int) ([]Node, Node, error) {
	params := make([]Node, 0, paramCount)

	n := firstNode
	var p Node
	var err error
	for ix := 0; ix < paramCount; ix++ {
		if n == nil {
			return nil, nil, errorNotEnoughParameters(firstNode)
		}

		switch nn := n.(type) {
		case *WordNode:
			p, n, err = evaluateNode(frame, nn)
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
