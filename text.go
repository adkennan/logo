package main

import (
	"fmt"
	"image/color"
	"unicode"
)

type KeyMessage struct {
	MessageBase
	Sym  uint32
	Mod  uint32
	Char rune
}

const (
	K_RETURN    = 0xd
	K_BACKSPACE = 0x8
	K_DELETE    = 0x7f
	K_UP        = 0x111
	K_DOWN      = 0x112
	K_RIGHT     = 0x113
	K_LEFT      = 0x114
	K_INSERT    = 0x115
	K_HOME      = 0x116
	K_END       = 0x117
	K_PAGEUP    = 0x118
	K_PAGEDOWN  = 0x119
	K_ESCAPE    = 0x1b

	K_NONE   = 0
	K_LSHIFT = 0x1
	K_RSHIFT = 0x2
	K_LCTRL  = 0x40
	K_RCTRL  = 0x80
	K_LALT   = 0x100
	K_RALT   = 0x200
	K_LMETA  = 0x400
	K_RMETA  = 0x800

	K_F1 = 0x11a
	K_F2 = 0x11b
	K_F3 = 0x11c
)

type ConsoleScreen struct {
	ws      *Workspace
	sfcs    [2]Surface
	sfcIx   int
	cx      int
	cy      int
	channel *Channel
}

func initConsole(workspace *Workspace, w, h int) *ConsoleScreen {
	sfcs := [2]Surface{
		workspace.screen.screen.CreateSurface(w, h, false),
		workspace.screen.screen.CreateSurface(w, h, false),
	}
	cs := &ConsoleScreen{
		workspace,
		sfcs,
		0,
		0,
		0,
		workspace.broker.Subscribe("Console", MT_KeyPress, MT_EditStart, MT_EditStop)}

	cs.sfcs[0].Clear()
	cs.sfcs[1].Clear()

	cs.sfcs[0].SetColor(color.RGBA{0, 0, 0, 255})
	cs.sfcs[0].Fill(0, 0, w, h)

	cs.sfcs[1].SetColor(color.RGBA{0, 0, 0, 255})
	cs.sfcs[1].Fill(0, 0, w, h)

	workspace.registerBuiltIn("CLEARTEXT", "CT", 0, _c_ClearText)
	workspace.registerBuiltIn("CURSOR", "", 0, _c_Cursor)
	workspace.registerBuiltIn("SETCURSOR", "", 1, _c_SetCursor)
	workspace.registerBuiltIn("WIDTH", "", 0, _c_Width)
	workspace.registerBuiltIn("HEIGHT", "", 1, _c_Height)

	return cs
}

func (this *ConsoleScreen) Id() int {
	return 0
}

func (this *ConsoleScreen) Name() string {
	return ""
}

func (this *ConsoleScreen) FirstLineOfSplitScreen() int {

	if this.cy <= splitScreenSize {
		return 0
	}
	return this.cy - splitScreenSize
}

func (this *ConsoleScreen) Clear() {

	s := this.sfcs[this.sfcIx]
	s.SetColor(color.RGBA{0, 0, 0, 255})
	s.Fill(0, 0, s.W(), s.H())

	this.cx = 0
	this.cy = 0

	this.channel.Publish(newRegionMessage(MT_UpdateText, s,
		[]*Region{&Region{0, 0, s.W(), s.H()}}))
}

func (this *ConsoleScreen) ReadChar() (rune, error) {

	this.channel.Resume()
	defer this.channel.Pause()
	m := this.channel.Wait()
	switch ks := m.(type) {
	case *KeyMessage:
		return ks.Char, nil
	}
	return 0, nil
}

func (this *ConsoleScreen) ReadLine() (string, error) {
	cursorPos := 0
	chars := make([]rune, 0, 10)
	this.drawEditLine(cursorPos, chars)
	this.channel.Resume()
	m := this.channel.Wait()
	for ; m != nil; m = this.channel.Wait() {
		switch ks := m.(type) {
		case *KeyMessage:
			{
				switch ks.Sym {
				case K_RETURN:
					line := string(chars)

					this.clearEditLine()
					this.Write(line)
					this.Write("\n")
					this.channel.Pause()
					return line, nil
				case K_LEFT:
					if cursorPos > 0 {
						cursorPos--
					}
				case K_RIGHT:
					if cursorPos < len(chars) {
						cursorPos++
					}
				case K_HOME:
					cursorPos = 0
				case K_END:
					cursorPos = len(chars)
				case K_BACKSPACE:
					if cursorPos > 0 {
						chars = append(chars[:cursorPos-1], chars[cursorPos:]...)
						cursorPos--
					}
				case K_DELETE:
					if cursorPos < len(chars) {
						chars = append(chars[:cursorPos], chars[cursorPos+1:]...)
					}
				default:
					if ks.Char != 0 {
						if unicode.IsGraphic(ks.Char) {
							if cursorPos == len(chars) {
								chars = append(chars, ks.Char)
							} else {
								chars = append(chars[:cursorPos],
									append([]rune{ks.Char}, chars[cursorPos:]...)...)
							}
							cursorPos++
						}
					}
				}
				this.drawEditLine(cursorPos, chars)
			}
		}
	}
	return "", nil
}

func (this *ConsoleScreen) clearEditLine() {
	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	dst := this.sfcs[this.sfcIx]

	dst.SetColor(color.RGBA{0, 0, 0, 255})
	dst.Fill(nx, ny, dst.W(), ny+gm.charHeight)
}

func (this *ConsoleScreen) drawEditLine(cursorPos int, chars []rune) {

	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	dst := this.sfcs[this.sfcIx]

	sx := nx
	sy := ny
	this.clearEditLine()

	maxX := gm.charWidth
	for _, c := range chars {
		nx = gm.renderGlyph(c, glyphStyleNormal, dst, nx, ny)
		if nx >= dst.W() {
			break
		}
		if nx > maxX {
			maxX = nx
		}
	}

	dst.SetColor(color.RGBA{255, 255, 255, 255})
	dst.Fill(2+(this.cx+cursorPos)*gm.charWidth,
		ny+gm.charHeight-2,
		((this.cx+cursorPos+1)*gm.charWidth)-2,
		ny+gm.charHeight)

	this.channel.Publish(newRegionMessage(MT_UpdateText, dst,
		[]*Region{&Region{sx, sy, maxX + gm.charWidth, gm.charHeight}}))
}

func (this *ConsoleScreen) Surface() Surface {
	return this.sfcs[this.sfcIx]
}

func (this *ConsoleScreen) Write(text string) error {

	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	sx := nx
	sy := ny
	ex := nx + gm.charWidth
	ey := ny + gm.charHeight
	for _, c := range text {
		if unicode.IsGraphic(c) {
			nx = gm.renderGlyph(c, glyphStyleNormal, this.sfcs[this.sfcIx], nx, ny)
			this.cx++
			if nx > ex {
				ex = nx
			}
		}
		if c == '\n' || nx >= this.sfcs[this.sfcIx].W() {
			this.cx = 0
			nx = 0
			this.cy++
			ny += gm.charHeight
			if ny+gm.charHeight >= this.sfcs[0].H() {

				src := this.sfcs[this.sfcIx]
				dst := this.sfcs[1-this.sfcIx]

				dst.SetColor(color.RGBA{0, 0, 0, 255})
				dst.Fill(0, 0, dst.W(), dst.H())

				dst.DrawSurface(0, -gm.charHeight, src)

				this.sfcIx = 1 - this.sfcIx

				this.cy--
				ny -= gm.charHeight

				sx = 0
				sy = 0
				ex = dst.W()
			}
			if ny > ey {
				ey = ny
			}
		}
	}

	this.channel.Publish(newRegionMessage(MT_UpdateText, this.sfcs[this.sfcIx],
		[]*Region{&Region{sx, sy, ex - sx, ey - sy}}))

	return nil
}

func (this *ConsoleScreen) Close() error {
	return nil
}

func (this *ConsoleScreen) IsInteractive() bool {
	return true
}

func _c_ClearText(frame Frame, parameters []Node) *CallResult {

	frame.workspace().console.Clear()

	return nil
}

func _c_Cursor(frame Frame, parameters []Node) *CallResult {

	c := frame.workspace().console

	x := newWordNode(-1, -1, fmt.Sprint(c.cx), true)
	y := newWordNode(-1, -1, fmt.Sprint(c.cy), true)
	x.addNode(y)
	return returnResult(newListNode(-1, -1, x))
}

func _c_SetCursor(frame Frame, parameters []Node) *CallResult {
	fx, fy, err := parseCoords(parameters[0])
	if err != nil {
		return errorResult(err)
	}
	x := int(fx)
	y := int(fy)

	c := frame.workspace().console
	gm := c.ws.glyphMap
	mx := c.sfcs[0].W() / gm.charWidth
	my := c.sfcs[0].H() / gm.charHeight
	if x < 0 || x > mx || y < 0 || y > my {
		ln, _ := parameters[0].(*ListNode)
		return errorResult(errorInvalidPosition(ln))
	}

	c.cx = x
	c.cy = y

	return nil
}

func _c_Width(frame Frame, parameters []Node) *CallResult {
	c := frame.workspace().console
	gm := c.ws.glyphMap
	mx := c.sfcs[0].W() / gm.charWidth

	return returnResult(newWordNode(-1, -1, fmt.Sprint(mx), true))
}

func _c_Height(frame Frame, parameters []Node) *CallResult {
	c := frame.workspace().console
	gm := c.ws.glyphMap
	my := c.sfcs[0].H() / gm.charHeight

	return returnResult(newWordNode(-1, -1, fmt.Sprint(my), true))
}
