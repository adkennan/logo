package main

import (
	"image/color"
	"unicode"
)

type KeyMessage struct {
	MessageBase
	Sym  uint32
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
		workspace.screen.screen.CreateSurface(w, h),
		workspace.screen.screen.CreateSurface(w, h),
	}
	cs := &ConsoleScreen{
		workspace,
		sfcs,
		0,
		0,
		0,
		workspace.broker.Subscribe(MT_KeyPress)}

	cs.sfcs[0].Clear()
	cs.sfcs[1].Clear()

	cs.sfcs[0].SetColor(color.RGBA{0, 0, 0, 255})
	cs.sfcs[0].Fill(0, 0, w, h)

	cs.sfcs[1].SetColor(color.RGBA{0, 0, 0, 255})
	cs.sfcs[1].Fill(0, 0, w, h)

	workspace.registerBuiltIn("CLEARTEXT", "CT", 0, _c_ClearText)

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

	this.channel.PublishId(MT_UpdateText)
}

func (this *ConsoleScreen) ReadLine() (string, error) {
	cursorPos := 0
	chars := make([]rune, 0, 10)
	this.drawEditLine(cursorPos, chars)

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
					if cursorPos < len(chars)-1 {
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

	this.clearEditLine()

	for _, c := range chars {
		nx = gm.renderGlyph(c, glyphStyleNormal, dst, nx, ny)
	}

	dst.SetColor(color.RGBA{255, 255, 255, 255})
	dst.Fill(this.cx+((cursorPos+2)*gm.charWidth)+2,
		ny+gm.charHeight-2,
		this.cx+((cursorPos+3)*gm.charWidth)-2,
		ny+gm.charHeight)

	this.channel.PublishId(MT_UpdateText)
}

func (this *ConsoleScreen) Surface() Surface {
	return this.sfcs[this.sfcIx]
}

func (this *ConsoleScreen) Write(text string) error {

	gm := this.ws.glyphMap
	nx := this.cx * gm.charWidth
	ny := this.cy * gm.charHeight

	for _, c := range text {
		if unicode.IsGraphic(c) {
			nx = gm.renderGlyph(c, glyphStyleNormal, this.sfcs[this.sfcIx], nx, ny)
			this.cx++

		} else if c == '\n' {
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
			}
		}
	}

	this.channel.PublishId(MT_UpdateText)

	return nil
}

func (this *ConsoleScreen) Close() error {
	return nil
}

func (this *ConsoleScreen) IsInteractive() bool {
	return true
}

func _c_ClearText(frame Frame, parameters []Node) (Node, error) {

	frame.workspace().console.Clear()

	return nil, nil
}
