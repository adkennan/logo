package logo

/*
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

func (this *InterpretedProcedure) createFrame(scope Scope, frame Frame, caller *WordNode) Frame {

	writeTrace(this.name, frame)
	return &InterpretedFrame{scope, frame, caller, this, nil, nil, make(map[string]*Variable), false}
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

func (this *InterpretedScope) addProcedure(proc *InterpretedProcedure) {
	this.procedures[proc.name] = proc
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

func CreateInterpretedScope(parentScope Scope) *InterpretedScope {
	procs := make(map[string]*InterpretedProcedure, 8)
	return &InterpretedScope{procs, nil, parentScope}
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

var promptPrimary = "? "
var promptSecondary = "> "
var greeting = "\nWelcome to Logo\n\n"

func RunInterpreter() error {

	biScope := CreateBuiltInScope()
	scope := CreateInterpretedScope(biScope)

	prompt := promptPrimary
	definingProc := false
	partial := ""
	SelectedFile().Write(greeting)

	for {
		f := SelectedFile()
		if f.IsInteractive() {
			f.Write(prompt)
		}
		line, err := SelectedFile().ReadLine()
		if err != nil {
			return err
		}
		lu := strings.ToUpper(line)

		if definingProc {
			partial += "\n" + line
			if lu == keywordEnd {
				fn, err := ParseString(partial)
				if err != nil {
					f.Write(err.Error())
					f.Write("\n")
				} else {
					proc, _, err := readInterpretedProcedure(fn)
					if err != nil {
						f.Write(err.Error())
						f.Write("\n")
					} else {
						scope.addProcedure(proc)
						f.Write(proc.name + " created.\n")
					}
					partial = ""
					prompt = promptPrimary
					definingProc = false
				}
			}
		} else {
			if line == "" {
				continue
			}
			if strings.HasPrefix(lu, keywordTo) {
				definingProc = true
				prompt = promptSecondary
				partial = line
			} else {
				if partial != "" {
					line = partial + "\n" + line
				}

				if strings.HasSuffix(lu, "~") {
					partial = line[0 : len(line)-1]
					prompt = promptSecondary
				} else {
					err = Evaluate(scope, line)
					partial = ""
					prompt = promptPrimary
					if err != nil {
						f.Write(err.Error())
						f.Write("\n")
					}
				}
			}
		}
	}
}
*/
