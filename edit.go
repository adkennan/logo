package main

import (
	"bytes"
	"fmt"
	"image/color"
	"strings"
	"unicode"
)

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

	img := ws.screen.screen.CreateSurface(w, h)

	c := ws.broker.Subscribe(MT_KeyPress)
	c.Pause()

	ws.registerBuiltIn("EDIT", "", 1, _ed_Edit)

	return &Editor{
		ws,
		c,
		nil,
		"",
		0,
		0, 0,
		0, 0, 0, 0,
		img,
		make(chan bool)}
}

func (this *Editor) StartEditor(content string) string {

	this.buffer = []string{}
	this.x = 0
	this.y = 0
	this.sx1 = 0
	this.sy1 = 0
	this.sx2 = 0
	this.sy2 = 0
	this.clip = ""
	this.image.Clear()

	this.Insert(content)

	this.y = 0

	go this.RunEditor()

	<-this.done

	return strings.Join(this.buffer, "\n")
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

	l := this.buffer[this.sy1][this.sx1:] + this.buffer[this.sy2][:this.sx2]
	this.buffer[this.sy1] = l
	linesDeleted = this.sy2 - this.sy1
	if this.sy2 > this.sy1 {
		this.buffer = append(this.buffer[:this.sy1+1], this.buffer[this.sy2:]...)
	}

	linesDeleted = this.sy2 - this.sy1

	this.sy1 = 0
	this.sx1 = 0
	this.sy2 = 0
	this.sx2 = 0

	return
}

func (this *Editor) CopySelection() {
	this.clip = this.GetSelection()
}

func (this *Editor) CutSelection() {
	this.CopySelection()
	this.DeleteSelection()
}

func (this *Editor) GetSelection() string {
	if this.HasSelection() {
		return ""
	}

	if this.sy1 == this.sy2 {
		return this.buffer[this.sy1][this.sx1:this.sx2]
	}

	var b bytes.Buffer
	lix := this.sy2 - this.sy2
	for ix, il := range this.buffer[this.sy1:this.sy2] {
		if ix == 0 {
			b.WriteString(this.buffer[this.sy1][this.sx1:])
		} else if ix == lix {
			b.WriteString(this.buffer[this.sy1+ix][:this.sx2])
		} else {
			b.WriteString(il)
		}
		b.WriteRune('\n')
	}
	return b.String()
}

func (this *Editor) InsertRune(c rune) (linesInserted int) {

	l := this.buffer[this.y]
	sl := l[:this.x]
	el := l[this.x:]
	if c == '\n' {

		this.buffer[this.y] = sl
		this.buffer = append(append(this.buffer[:this.y], el), this.buffer[this.y:]...)
		this.y++
		this.x = 0
		return 1
	} else {
		this.buffer[this.y] = fmt.Sprint(el, string(c), sl)
		this.x++
		return 0
	}
}

func (this *Editor) DeleteRune() (linesDeleted int) {

	l := this.buffer[this.y]
	if this.x == len(l) {
		if this.y < len(this.buffer) {
			this.buffer[this.y] = this.buffer[this.y+1]
			this.buffer = append(this.buffer[:this.y], this.buffer[this.y+1:]...)
			return 1
		}
	} else {
		this.buffer[this.y] = l[:this.x-1] + l[this.x:]
	}
	return 0
}

func (this *Editor) HasSelection() bool {
	return this.sy1 != this.sy2 || this.sx1 != this.sx2
}

func (this *Editor) IsInSelection(x, y int) bool {
	if !this.HasSelection() {
		return false
	}

	if y == this.sy1 {
		return x >= this.sx1
	}

	if y == this.sy2 {
		return x <= this.sx2
	}

	return this.sy1 < y && y < this.sy2
}

func (this *Editor) RenderLines(s, e int) {

	gm := this.ws.glyphMap
	gh := gm.charHeight
	gw := gm.charWidth
	lw := this.image.W() / gw
	lc := this.image.H() / gh

	if s > this.vy+lc || e < this.vy {
		return
	}

	cy := s
	if s < this.vy {
		s = this.vy
	}
	if e > this.vy+lc {
		e = this.vy + lc
	}

	if e > len(this.buffer) {
		e = len(this.buffer)
	}

	y := (s - this.vy) * gh
	sy := y

	for _, l := range this.buffer[s:e] {
		x := 0
		for cx, c := range l {
			gs := glyphStyleNormal
			if this.IsInSelection(cx, cy) {
				gs = glyphStyleInverse
			}
			x = gm.renderGlyph(c, gs, this.image, x, y)
			if x >= lw {
				x = 0
				y += gh
			}
		}
		cy++
		y += gh
	}

	this.image.SetColor(color.RGBA{255, 255, 255, 255})
	this.image.Fill((this.x*gm.charWidth)+2,
		((this.y+1)*gm.charHeight)-2,
		((this.x+1)*gm.charWidth)-2,
		((this.y + 1) * gm.charHeight))

	this.channel.Publish(newRegionMessage(MT_UpdateEdit, this.image, []*Region{&Region{0, sy, this.image.W(), y}}))
}

func (this *Editor) CursorLeft() bool {
	if this.x > 0 {
		this.x--
		return true
	} else if this.y > 1 {
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
	} else if this.y < len(this.buffer) {
		this.y++
		this.x = 0
		return true
	}
	return false
}

func (this *Editor) CursorUp() bool {
	if this.y > 1 {
		this.y--
		if this.x > len(this.buffer[this.y]) {
			this.x = len(this.buffer[this.y])
		}
		return true
	}
	return false
}

func (this *Editor) CursorDown() bool {
	if this.y < len(this.buffer) {
		this.y++
		if this.x > len(this.buffer[this.y]) {
			this.x = len(this.buffer[this.y])
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

	this.image.Clear()
	this.channel.Publish(newRegionMessage(MT_UpdateEdit, this.image, []*Region{&Region{0, 0, this.image.W(), this.image.H()}}))
	this.RenderLines(0, len(this.buffer))

	defer func() {
		this.channel.Pause()
		this.channel.PublishId(MT_EditStop)
		this.done <- true
	}()

	exit := false
	m := this.channel.Wait()
	for ; m != nil && !exit; m = this.channel.Wait() {
		sl := this.y
		el := sl + 1
		switch ks := m.(type) {
		case *KeyMessage:
			{
				switch ks.Sym {
				case K_LEFT:
					this.CursorLeft()
				case K_RIGHT:
					this.CursorRight()
				case K_UP:
					this.CursorUp()
				case K_DOWN:
					this.CursorDown()
				case K_HOME:
					this.CursorHome()
				case K_END:
					this.CursorEnd()
				case K_BACKSPACE:
					if this.CursorLeft() {
						if this.DeleteRune() > 0 {
							el = len(this.buffer)
						}
					}
				case K_DELETE:
					if this.DeleteRune() > 0 {
						el = len(this.buffer)
					}
				case K_ESCAPE:
					exit = true
				default:
					if ks.Char != 0 && unicode.IsGraphic(ks.Char) {
						if this.InsertRune(ks.Char) > 0 {
							el = len(this.buffer)
						}
					}
				}
				if this.y < sl {
					sl = this.y
				}
				this.RenderLines(sl, el)
			}
		}
	}

}

func _ed_Edit(frame Frame, parameters []Node) (Node, error) {

	names, err := toWordList(parameters[0])
	if err != nil {
		return nil, err
	}
	ws := frame.workspace()
	var b bytes.Buffer

	for _, n := range names {
		p := ws.findProcedure(strings.ToUpper(n.value))
		if p == nil {
			continue
		}
		switch ip := p.(type) {
		case *InterpretedProcedure:
			b.WriteString(ip.source)
			b.WriteString("\n\n")
		}
	}

	editedContent := ws.editor.StartEditor("a\n b\n  c\n  d\n") //b.String())

	print(editedContent)

	return nil, nil
}
