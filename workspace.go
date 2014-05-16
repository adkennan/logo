package main

import (
	"bufio"
	"bytes"
	"os"
	"os/user"
	"path"
	"strings"
)

var promptPrimary = "? "
var promptSecondary = "> "
var greeting = "\nWelcome to Logo\n\n"

type Workspace struct {
	rootFrame    *RootFrame
	procedures   map[string]Procedure
	traceEnabled bool
	broker       *MessageBroker
	files        *Files
	screen       *Screen
	turtle       *Turtle
	glyphMap     *GlyphMap
	console      *ConsoleScreen
	editor       *Editor
}

func CreateWorkspace() *Workspace {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	ws := &Workspace{nil, make(map[string]Procedure, 100), false, nil, nil, nil, nil, nil, nil, nil}
	ws.rootFrame = &RootFrame{ws, nil, nil, make(map[string]*Variable, 10)}
	ws.broker = CreateMessageBroker()
	ws.files = CreateFiles(path.Join(u.HomeDir, "logo"))
	registerBuiltInProcedures(ws)

	ws.screen = initScreen(ws)
	ws.turtle = initTurtle(ws)
	ws.glyphMap = initGlyphMap()
	ws.console = initConsole(ws, ws.screen.screen.W(), ws.screen.screen.H())
	ws.editor = initEditor(ws, ws.screen.screen.W(), ws.screen.screen.H())

	ws.files.defaultFile = ws.console
	ws.files.writer = ws.console
	ws.files.reader = ws.console

	ws.screen.Open()

	return ws
}

func (this *Workspace) Screen() *Screen { return this.screen }

func (this *Workspace) exit() {
	os.Exit(0)
}

func (this *Workspace) print(text string) error {

	err := this.files.writer.Write(text)
	return err
}

func (this *Workspace) setTrace(trace bool) {
	this.traceEnabled = trace
}

func (this *Workspace) trace(depth int, text string) error {

	if !this.traceEnabled {
		return nil
	}
	err := this.files.writer.Write(strings.Repeat(" ", depth) + text)
	return err
}
func (this *Workspace) addProcedure(proc *InterpretedProcedure) {
	this.procedures[proc.name] = proc
}

func (this *Workspace) findProcedure(name string) Procedure {

	p, _ := this.procedures[name]

	return p
}

func (this *Workspace) registerBuiltIn(longName, shortName string, paramCount int, f evaluator) {
	p := &BuiltInProcedure{longName, paramCount, f}

	this.procedures[longName] = p
	if shortName != "" {
		this.procedures[shortName] = p
	}
}

func (this *Workspace) evaluate(source string) error {

	n, err := ParseString(source)
	if err != nil {
		return err
	}

	this.rootFrame.node = n
	defer func() {
		this.rootFrame.node = nil
	}()

	n, err = this.rootFrame.eval(make([]Node, 0, 0))
	if err != nil {
		return err
	}
	return nil
}

func (this *Workspace) RunInterpreter() {
	this.print(greeting)

	go this.readFile()

	l := this.broker.Subscribe(MT_Quit)
	l.Wait()
}

func (this *Workspace) readString(text string) error {
	b := bytes.NewBufferString(text)
	s := bufio.NewScanner(b)

	fw := this.files.writer
	partial := ""
	definingProc := false
	for s.Scan() {
		line := s.Text()
		lu := strings.ToUpper(line)

		if definingProc {
			partial += "\n" + line
			if lu == keywordEnd {
				fn, err := ParseString(partial)
				if err != nil {
					return err
				} else {
					proc, _, err := readInterpretedProcedure(fn)
					if err != nil {
						return err
					} else {
						proc.source = partial
						this.addProcedure(proc)
						fw.Write(proc.name + " defined.\n")
					}
					partial = ""
					definingProc = false
				}
			}
		} else {
			if line == "" {
				continue
			}
			if strings.HasPrefix(lu, keywordTo) {
				definingProc = true
				partial = line
			} else {

				err := this.evaluate(line)
				if err != nil {
					return err
				}
			}
		}
	}

	return s.Err()
}

func (this *Workspace) readFile() error {
	prompt := promptPrimary
	definingProc := false
	partial := ""

	for {
		fw := this.files.writer
		fr := this.files.reader
		if fr.IsInteractive() {
			fw.Write(prompt)
		}
		line, err := fr.ReadLine()
		if err != nil {
			return err
		}
		lu := strings.ToUpper(line)

		if definingProc {
			partial += "\n" + line
			if lu == keywordEnd {
				fn, err := ParseString(partial)
				if err != nil {
					fw.Write(err.Error())
					fw.Write("\n")
				} else {
					proc, _, err := readInterpretedProcedure(fn)
					if err != nil {
						fw.Write(err.Error())
						fw.Write("\n")
					} else {
						proc.source = partial
						this.addProcedure(proc)
						fw.Write(proc.name + " defined.\n")
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
					err = this.evaluate(line)
					partial = ""
					prompt = promptPrimary
					if err != nil {
						fw.Write(err.Error())
						fw.Write("\n")
					}
				}
			}
		}
	}
}
