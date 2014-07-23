package main

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const noSelection = -1

type Editor struct {
	ws                 *Workspace
	channel            *Channel
	buffer             []string
	clip               string
	vy                 int
	x, y               int
	sx1, sy1, sx2, sy2 int
	image              Surface
	done               chan bool
}

func initEditor(ws *Workspace, w, h int) *Editor {

	img := ws.screen.screen.CreateSurface(w, h, false)

	c := ws.broker.Subscribe("Editor", MT_KeyPress)
	c.Pause()

	ws.registerBuiltIn(keywordEdit, "", 0, _ed_Edit)

	e := &Editor{
		ws,
		c,
		nil,
		"",
		0,
		0, 0,
		0, 0, 0, 0,
		img,
		make(chan bool)}

	return e
}

func (this *Editor) initEditorUi() {

	gm := this.ws.glyphMap
	gh := gm.charHeight
	//gw := gm.charWidth

	this.image.Clear()
	this.image.SetColor(colorWhite)
	this.image.DrawLine(0, this.image.H()-(gh*2), this.image.W(), this.image.H()-(gh*2))

	this.writeStatus(0, "F1: Save    ESC: Cancel")

	this.invalidate(0, 0, this.image.W(), this.image.H())
}

func (this *Editor) writeStatus(x int, text string) {

	gm := this.ws.glyphMap
	y := this.image.H() - ((gm.charHeight * 3) / 2)

	x2 := x
	for _, c := range text {
		x2 = gm.renderGlyph(c, glyphStyleNormal, this.image, x2, y)
	}

	this.invalidate(x, y, x2, y+gm.charHeight)
}

func (this *Editor) StartEditor(content string) string {

	this.buffer = []string{}
	this.x = 0
	this.y = 0
	this.sx1 = noSelection
	this.sy1 = noSelection
	this.sx2 = noSelection
	this.sy2 = noSelection
	this.clip = ""

	this.initEditorUi()

	this.Insert(content)

	this.y = 0

	go this.RunEditor()

	r := <-this.done

	if r {
		return strings.Join(this.buffer, "\n")
	}
	return ""
}

func (this *Editor) Insert(content string) (linesInserted int) {

	lines := strings.Split(content, "\n")

	var l string
	if len(this.buffer) > 0 {
		l = this.buffer[this.y]
	}
	ll := len(lines)
	sl := l[:this.x]
	el := l[this.x:]
	lines[0] = sl + lines[0]
	lines[ll-1] = lines[ll-1] + el
	this.buffer = append(this.buffer[:this.y], append(make([]string, ll, ll), this.buffer[this.y:]...)...)

	for _, il := range lines {
		this.buffer[this.y] = il
		this.y++
	}

	return len(lines) - 1
}

func (this *Editor) DeleteSelection() (linesDeleted int) {

	s := intMin(this.sy1, this.sy2)
	e := intMax(this.sy1, this.sy2)
	sl := this.sx1
	el := this.sx2
	if this.sy1 == this.sy2 {
		sl = intMin(this.sx1, this.sx2)
		el = intMax(this.sx1, this.sx2)
	} else if this.sy1 > this.sy2 {
		sl = this.sx2
		el = this.sx1
	}

	linesDeleted = e - s
	start := this.buffer[s][:sl]
	end := this.buffer[e][intMin(len(this.buffer[e]), el+1):]
	l := start + end
	this.buffer[s] = l
	if e > s {
		this.buffer = append(this.buffer[:s+1], this.buffer[e+1:]...)
	}

	this.x = sl
	this.y = s
	this.ClearSelection()

	return
}

func (this *Editor) CopySelection() {
	this.clip = this.GetSelection()
}

func (this *Editor) CutSelection() {
	this.CopySelection()
	this.DeleteSelection()
}

func (this *Editor) PasteSelection() int {
	if len(this.clip) == 0 {
		return 0
	}
	linesInserted := 0
	for _, c := range this.clip {
		linesInserted += this.InsertRune(c)
	}
	return linesInserted
}

func (this *Editor) GetSelection() string {
	if !this.HasSelection() {
		return ""
	}

	if this.sy1 == this.sy2 {
		lb := this.buffer[this.sy1]
		e := intMax(this.sx1, this.sx2) + 1
		if e > len(lb) {
			e = len(lb)
		}
		return lb[intMin(this.sx1, this.sx2):e]
	}

	var b bytes.Buffer
	s := intMin(this.sy1, this.sy2)
	e := intMax(this.sy1, this.sy2)
	sl := this.sx1
	el := this.sx2
	if this.sy1 == this.sy2 {
		sl = intMin(this.sx1, this.sx2)
		el = intMax(this.sx1, this.sx2)
	} else if this.sy1 > this.sy2 {
		sl = this.sx2
		el = this.sx1
	}
	lix := e - s
	for ix, il := range this.buffer[s : e+1] {
		if ix == 0 {
			b.WriteString(this.buffer[s][sl:])
			b.WriteRune('\n')
		} else if ix == lix {
			e := el + 1
			lb := this.buffer[s+ix]
			if e > len(lb) {
				e = len(lb)
			}
			b.WriteString(lb[:e])
		} else {
			b.WriteString(il)
			b.WriteRune('\n')
		}
	}
	return b.String()
}

func (this *Editor) InsertRune(c rune) (linesInserted int) {

	l := this.buffer[this.y]
	if len(l) == 0 {
		if c == '\n' {
			this.buffer = append(this.buffer, "")
			copy(this.buffer[this.y+1:], this.buffer[this.y:])
			this.y++
			this.x = 0
			return 1
		}
		this.buffer[this.y] = string(c)
		this.x++
	} else {
		sl := l[:this.x]
		el := l[this.x:]
		if c == '\n' {

			this.buffer = append(this.buffer, "")
			copy(this.buffer[this.y+1:], this.buffer[this.y:])
			this.buffer[this.y] = sl
			this.buffer[this.y+1] = el
			this.y++
			this.x = 0
			return 1
		} else {
			this.buffer[this.y] = fmt.Sprint(sl, string(c), el)
			this.x++
		}
	}
	return 0
}

func (this *Editor) DeleteRune() (linesDeleted int) {

	l := this.buffer[this.y]
	if this.x == len(l) {
		if this.y < len(this.buffer)-1 {
			this.buffer[this.y] = this.buffer[this.y] + this.buffer[this.y+1]
			this.buffer = append(this.buffer[:this.y+1], this.buffer[this.y+2:]...)
			return 1
		}
	} else {
		this.buffer[this.y] = l[:this.x] + l[this.x+1:]
	}
	return 0
}

func (this *Editor) HasSelection() bool {
	return this.sx1 != noSelection
}

func (this *Editor) IsInSelection(x, y int) bool {
	if !this.HasSelection() {
		return false
	}

	if this.sy1 == this.sy2 {
		if y != this.sy1 {
			return false
		}
		x1, x2 := sortInts(this.sx1, this.sx2)
		return x1 <= x && x <= x2

	} else if this.sy1 < this.sy2 {
		if y == this.sy1 {
			return x >= this.sx1
		}
		if y == this.sy2 {
			return x <= this.sx2
		}
		return this.sy1 < y && y < this.sy2
	}

	if y == this.sy2 {
		return x >= this.sx2
	}
	if y == this.sy1 {
		return x <= this.sx1
	}
	return this.sy2 < y && y < this.sy1
}

func (this *Editor) RenderLines(s, e int) {

	gm := this.ws.glyphMap
	gh := gm.charHeight
	gw := gm.charWidth
	lw := (this.image.W() - 2) / gw
	lc := (this.image.H() - 2) / gh

	if s > this.vy+lc || e < this.vy {
		return
	}

	if s < this.vy {
		s = this.vy
	}
	cy := s

	if e > this.vy+lc {
		e = this.vy + lc
	}

	if e > len(this.buffer) {
		e = len(this.buffer)
	}

	y := (s - this.vy) * gh
	sy := y

	this.image.SetColor(colorBlack)
	this.image.Fill(0, y, this.image.W(), (e-this.vy)*gh)

	for _, l := range this.buffer[s:e] {
		x := gw
		for cx, c := range l {
			gs := glyphStyleNormal
			if this.IsInSelection(cx, cy) {
				gs = glyphStyleInverse
			}
			x = gm.renderGlyph(c, gs, this.image, x, y)
			if cx >= lw {
				x = gw
				y += gh
			}
		}
		cy++
		y += gh
	}

	this.image.SetColor(colorWhite)
	this.image.Fill(((this.x+1)*gm.charWidth)+2,
		((this.y+1)*gm.charHeight)-2,
		((this.x+2)*gm.charWidth)-2,
		((this.y+1)*gm.charHeight)-1)

	this.invalidate(0, sy, this.image.W(), y)

	this.writeStatus(this.image.W()/2, fmt.Sprintf("Line: %d Col: %d    ", this.x, this.y))
}

func (this *Editor) invalidate(x, y, w, h int) {
	this.channel.Publish(newRegionMessage(MT_UpdateEdit, this.image, []*Region{&Region{x, y, w, h}}))
}

func (this *Editor) CursorLeft() bool {
	if this.x > 0 {
		this.x--
		return true
	} else if this.y > 0 {
		this.y--
		this.x = len(this.buffer[this.y])
		return true
	}
	return false
}

func (this *Editor) CursorRight() bool {
	if this.x < len(this.buffer[this.y]) {
		this.x++
		return true
	} else if this.y < len(this.buffer)-1 {
		this.y++
		this.x = 0
		return true
	}
	return false
}

func (this *Editor) CursorUp() bool {
	if this.y > 0 {
		this.y--
		if this.x >= len(this.buffer[this.y]) {
			this.x = intMax(0, len(this.buffer[this.y]))
		}
		return true
	}
	return false
}

func (this *Editor) CursorDown() bool {
	if this.y < len(this.buffer)-1 {
		this.y++
		if this.x >= len(this.buffer[this.y]) {
			this.x = intMax(0, len(this.buffer[this.y]))
		}
		return true
	}
	return false
}

func (this *Editor) CursorHome() {
	this.x = 0
}

func (this *Editor) CursorEnd() {
	this.x = len(this.buffer[this.y])
}

func (this *Editor) RunEditor() {

	this.channel.Resume()
	this.channel.PublishId(MT_EditStart)

	this.channel.Publish(newRegionMessage(MT_UpdateEdit, this.image, []*Region{&Region{0, 0, this.image.W(), this.image.H()}}))
	this.RenderLines(0, len(this.buffer))

	shouldSave := false
	defer func() {
		this.channel.Pause()
		this.channel.PublishId(MT_EditStop)
		this.done <- shouldSave
	}()

	exit := false
	m := this.channel.Wait()
	for ; m != nil && !exit; m = this.channel.Wait() {
		sl := this.y
		el := -1
		switch ks := m.(type) {
		case *KeyMessage:
			{
				pcx := this.x
				pcy := this.y
				updateSel := false
				shiftHeld := (ks.Mod&K_LSHIFT) == K_LSHIFT || (ks.Mod&K_RSHIFT) == K_RSHIFT
				ctrlHeld := (ks.Mod&K_LCTRL) == K_LCTRL || (ks.Mod&K_RCTRL) == K_RCTRL

				if ctrlHeld {
					switch ks.Sym {
					case 'c':
						if this.HasSelection() {
							this.CopySelection()
						}
					case 'x':
						if this.HasSelection() {
							this.CutSelection()
							el = len(this.buffer)
						}
					case 'v':
						this.PasteSelection()
						el = len(this.buffer)
					}
				} else {
					switch ks.Sym {
					case K_LEFT:
						this.CursorLeft()
						updateSel = true
					case K_RIGHT:
						this.CursorRight()
						updateSel = true
					case K_UP:
						this.CursorUp()
						updateSel = true
					case K_DOWN:
						this.CursorDown()
						updateSel = true
					case K_HOME:
						this.CursorHome()
						updateSel = true
					case K_END:
						this.CursorEnd()
						updateSel = true
					case K_BACKSPACE:
						if this.CursorLeft() {
							if this.HasSelection() {
								this.DeleteSelection()
								el = len(this.buffer)
							} else if this.DeleteRune() > 0 {
								el = len(this.buffer)
								this.ClearSelection()
							}
						}
					case K_DELETE:
						if this.HasSelection() {
							this.DeleteSelection()
							el = len(this.buffer)
						} else if this.DeleteRune() > 0 {
							el = len(this.buffer)
							this.ClearSelection()
						}
					case K_ESCAPE:
						exit = true
					case K_F1:
						exit = true
						shouldSave = true
					case K_RETURN:
						if this.HasSelection() {
							this.DeleteSelection()
						}
						this.InsertRune('\n')
						el = len(this.buffer)

					default:
						if ks.Char != 0 && unicode.IsGraphic(ks.Char) {
							if this.HasSelection() {
								this.DeleteSelection()
							}
							if this.InsertRune(ks.Char) > 0 {
								this.ClearSelection()
								el = len(this.buffer)
							}
						}
					}
					if updateSel {
						if shiftHeld {
							if this.sx1 == noSelection {
								this.sx1 = pcx
								this.sy1 = pcy
							}
							this.sx2 = this.x
							this.sy2 = this.y
						} else if ks.Mod == 0 && this.sx1 != noSelection {
							sl = intMin(sl, intMin(this.sy1, this.sy2))
							el = intMax(el, intMax(this.sy1, this.sy2))
							this.ClearSelection()
						}
					}
				}
				this.RenderLines(intMin(sl, this.y)-1, intMax(intMax(sl, el), this.y)+1)
			}
		}
	}
}

func (this *Editor) ClearSelection() {

	this.sx1 = noSelection
	this.sy1 = noSelection
	this.sx2 = noSelection
	this.sy2 = noSelection
}

func sortInts(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func _ed_Edit(frame Frame, parameters []Node) *CallResult {

	var b bytes.Buffer
	ws := frame.workspace()

	if len(parameters) > 0 {
		names, err := toWordList(parameters[0])
		if err != nil {
			return errorResult(err)
		}

		for _, n := range names {
			p := ws.findProcedure(strings.ToUpper(n.value))
			if p == nil {
				b.WriteString("TO " + n.value + "\n\nEND\n\n")
			}
			switch ip := p.(type) {
			case *InterpretedProcedure:
				b.WriteString(ip.source)
				b.WriteString("\n\n")
			}
		}
	}

	editedContent := ws.editor.StartEditor(b.String())

	err := frame.workspace().readString(editedContent)
	if err != nil {
		return errorResult(err)
	}
	return nil
}
