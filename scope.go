package logo

import (
	"strings"
)

type Scope interface {
	setParent(parent Scope)
	findProcedure(name string) Procedure
}

type InterpretedProcedure struct {
	name       string
	parameters []string
	firstNode  Node
}

func writeTrace(procName string, parentFrame Frame) {

	if !traceEnabled {
		return
	}

	depth := 0
	for f := parentFrame; f != nil; f = f.parentFrame() {
		depth++
	}

	if TraceHandler == nil {
		Print("> " + strings.Repeat(" ", depth) + procName + "\n")
	} else {
		TraceHandler(depth, procName)
	}
}

func (this *InterpretedProcedure) createFrame(scope Scope, frame Frame) Frame {

	writeTrace(this.name, frame)
	return &InterpretedFrame{scope, frame, this, nil, nil, make(map[string]*Variable)}
}

func (this *InterpretedProcedure) parameterCount() int {
	return len(this.parameters)
}

type InterpretedScope struct {
	procedures  map[string]*InterpretedProcedure
	firstNode   Node
	parentScope Scope
}

func (this *InterpretedScope) setParent(parent Scope) {
	this.parentScope = parent
}

func (this *InterpretedScope) findProcedure(name string) Procedure {
	p, _ := this.procedures[name]
	if p != nil {
		return p
	}

	if this.parentScope == nil {
		return nil
	}

	return this.parentScope.findProcedure(name)
}

func ParseNonInteractiveScope(parentScope Scope, source string) (Scope, error) {

	firstNode, err := ParseString(source)
	if err != nil {
		return nil, err
	}

	procs := make(map[string]*InterpretedProcedure, 8)
	var p *InterpretedProcedure
	n := firstNode
	for n != nil {
		p, n, err = readInterpretedProcedure(n)
		if err != nil {
			return nil, err
		}
		procs[p.name] = p
	}

	return &InterpretedScope{procs, firstNode, parentScope}, nil
}

type BuiltInScope struct {
	procedures  map[string]Procedure
	parentScope Scope
}

func (this *BuiltInScope) setParent(parentScope Scope) {
	this.parentScope = parentScope
}

func (this *BuiltInScope) findProcedure(name string) Procedure {

	p, _ := this.procedures[name]
	if p != nil {
		return p
	}

	if this.parentScope == nil {
		return nil
	}

	return this.parentScope.findProcedure(name)
}

func (this *BuiltInScope) registerBuiltIn(longName, shortName string, paramCount int, f evaluator) {
	p := &BuiltInProcedure{longName, paramCount, f}

	this.procedures[longName] = p
	if shortName != "" {
		this.procedures[shortName] = p
	}
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

	return &InterpretedProcedure{strings.ToUpper(procName), params, firstNode}, n.next(), nil
}
