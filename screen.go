package main

import "fmt"

const (
	screenModeText = iota
	screenModeSplit
	screenModeGraphic
	screenModeEdit
)

const splitScreenSize = 4

type Screen struct {
	screen     Window
	w, h       int
	ws         *Workspace
	screenMode int
	channel    *Channel
}

func initScreen(workspace *Workspace, w, h int) *Screen {

	ss := newWindow(workspace.broker, w, h)
	w = ss.W()
	h = ss.H()

	s := &Screen{ss, w, h, workspace, screenModeSplit, workspace.broker.Subscribe("Screen")}

	workspace.registerBuiltIn("FULLSCREEN", "FS", 0, _s_Fullscreen)
	workspace.registerBuiltIn("TEXTSCREEN", "TS", 0, _s_Textscreen)
	workspace.registerBuiltIn("SPLITSCREEN", "SS", 0, _s_Splitscreen)

	return s
}

func (this *Screen) Open() {
	this.setScreenMode(screenModeSplit)
	go this.Update()
}

func (this *Screen) setScreenMode(screenMode int) {
	this.screenMode = screenMode
	switch this.screenMode {
	case screenModeSplit:
		this.Invalidate(MT_UpdateGfx)
		this.Invalidate(MT_UpdateText)
		h := this.screen.H() - (this.ws.glyphMap.charHeight * splitScreenSize)
		this.channel.Publish(newVisibleAreaChangeMessage(this.screen.W(), h))

	case screenModeGraphic:

		this.Invalidate(MT_UpdateGfx)
		this.channel.Publish(newVisibleAreaChangeMessage(this.screen.W(), this.screen.H()))
	case screenModeText:

		this.Invalidate(MT_UpdateText)
	}
}

func (this *Screen) Update() {

	gm := this.ws.glyphMap
	t := this.ws.turtle

	prevSplitScreenLoc := -1

	for {
		screenDirty := false
		drawTurtle := false
		for m := this.channel.Wait(); m != nil; m = this.channel.Poll() {
			switch rm := m.(type) {
			case *MessageBase:
				{
					switch m.MessageType() {
					case MT_EditStart:
						this.setScreenMode(screenModeEdit)
						screenDirty = true
					case MT_EditStop:
						this.setScreenMode(screenModeSplit)
						screenDirty = true
					}
				}
			case *KeyMessage:
				if this.screenMode != screenModeEdit {
					switch rm.Sym {
					case K_F1:
						this.setScreenMode(screenModeSplit)
						screenDirty = true
					case K_F2:
						this.setScreenMode(screenModeGraphic)
						screenDirty = true
					case K_F3:
						this.setScreenMode(screenModeText)
						screenDirty = true
					}
				}
			case *RegionMessage:
				{
					switch m.MessageType() {
					case MT_UpdateEdit:
						{
							if this.screenMode != screenModeEdit {
								continue
							}
							for _, r := range rm.regions {
								this.screen.ClearRect(t.screenColor, r.x, r.y, r.w, r.h)
								this.screen.DrawSurfacePart(r.x, r.y, rm.surface, r.x, r.y, r.w, r.h)
							}
							screenDirty = true
						}
					case MT_UpdateGfx:
						{
							if this.screenMode == screenModeText {
								continue
							}

							if this.screenMode == screenModeSplit {
								th := gm.charHeight * splitScreenSize

								this.screen.SetClipRect(0, 0, this.w, this.h-th)
							}

							for _, r := range rm.regions {
								this.screen.ClearRect(t.screenColor, r.x, r.y, r.w, r.h)
								this.screen.DrawSurfacePart(r.x, r.y, rm.surface, r.x, r.y, r.w, r.h)
							}
							if t.turtleState == turtleStateShown {
								drawTurtle = true
							}

							this.screen.ClearClipRect()
							screenDirty = true
						}

					case MT_UpdateText:
						{

							gm := this.ws.glyphMap
							c := this.ws.console
							cs := c.Surface()
							switch this.screenMode {
							case screenModeText:
								for _, r := range rm.regions {
									this.screen.ClearRect(t.screenColor, r.x, r.y, r.w, r.h)
									this.screen.DrawSurfacePart(r.x, r.y, cs, r.x, r.y, r.w, r.h)
								}
							case screenModeSplit:
								th := gm.charHeight * splitScreenSize
								firstLine := c.FirstLineOfSplitScreen()
								fl := (1 + firstLine) * gm.charHeight
								this.screen.SetClipRect(0, this.h-th, this.w, this.h)
								if prevSplitScreenLoc != firstLine {

									this.screen.DrawSurfacePart(0, this.h-th, cs,
										0, fl, this.w, th)

									prevSplitScreenLoc = firstLine
								} else {
									ts := this.h - th

									for _, r := range rm.regions {
										this.screen.ClearRect(t.screenColor, r.x, ts+r.y, r.w, r.h)
										this.screen.DrawSurfacePart(r.x, ts+r.y-fl, cs, r.x, r.y, r.w, r.h)
									}
								}
								this.screen.ClearClipRect()
							}
							screenDirty = true
						}
					}
				}
			}
			if screenDirty {
				if drawTurtle {
					if this.screenMode == screenModeSplit {
						th := gm.charHeight * splitScreenSize
						this.screen.SetClipRect(0, 0, this.w, this.h-th)
					}
					this.DrawTurtle()
					this.screen.ClearClipRect()
				}
				this.screen.Update()
			}
		}
	}
}

func (this *Screen) DrawTurtle() {
	t := this.ws.turtle

	if t.spriteNeedsUpdate() {
		t.updateSprite()
	}

	x := int(t.x+float64(this.w/2)) - turtleSize
	y := int(-t.y+float64(this.h/2)) - turtleSize

	this.screen.DrawSurface(x, y, t.sprite)
}

func (this *Screen) Invalidate(msgId int) {
	r := &Region{0, 0, this.w - 1, this.h - 1}
	var sfc Surface
	if msgId == MT_UpdateGfx {
		sfc = this.ws.turtle.image
	} else {
		sfc = this.ws.console.Surface()
	}
	this.channel.Publish(newRegionMessage(msgId, sfc, []*Region{r}))
}

type Region struct {
	x, y, w, h int
}

func (this *Region) Area() int {
	return this.w * this.h
}

func intMin(n1, n2 int) int {
	if n1 < n2 {
		return n1
	}
	return n2
}

func intMax(n1, n2 int) int {
	if n1 < n2 {
		return n2
	}
	return n1
}

func (this *Region) CombinedArea(other *Region) int {

	w := intMax(this.x, other.x) - intMin(this.x, other.x)
	h := intMax(this.y, other.y) - intMin(this.y, other.y)

	return w * h
}

func (this *Region) Contains(other *Region) bool {
	return this.x < other.x && this.y < other.y &&
		this.x+this.w > other.x+other.w &&
		this.y+this.h > other.y+other.h
}

func (this *Region) Combine(other *Region) {

	x1 := intMin(this.x, other.x)
	y1 := intMin(this.y, other.y)
	x2 := intMax(this.x+this.w, other.x+other.w)
	y2 := intMax(this.y+this.h, other.y+other.h)

	this.x = x1
	this.y = y1
	this.w = x2 - x1
	this.h = y2 - y1
}

func (this *Region) ContainsPoint(x, y int) bool {
	return this.x <= x && this.y <= y &&
		this.x+this.w > x && this.y+this.h > y
}

func (this *Region) AdjacentTo(x, y int) bool {

	return (x == this.x-1 && this.y <= y && y < this.y+this.h) ||
		(x == this.x+this.w && this.y <= y && y < this.y+this.h) ||
		(y == this.y-1 && this.x <= x && x < this.x+this.w) ||
		(y == this.y+this.h && this.x <= x && x < this.x+this.w)
}

func (this *Region) ExpandToInclude(x, y int) {

	if x < this.x {
		this.x = x
	} else if x >= this.x+this.w {
		this.w = (x - this.x) + 1
	}

	if y < this.y {
		this.y = y
	} else if y >= this.y+this.h {
		this.h = (y - this.y) + 1
	}
}

func (this *Region) Overlaps(other *Region) bool {
	if this.ContainsPoint(other.x, other.y) ||
		this.ContainsPoint(other.x+other.w, other.y) ||
		this.ContainsPoint(other.x+other.w, other.y+other.h) ||
		this.ContainsPoint(other.x, other.y+other.h) ||
		other.ContainsPoint(this.x, this.y) ||
		other.ContainsPoint(this.x+this.w, this.y) ||
		other.ContainsPoint(this.x+this.w, this.y+this.h) ||
		other.ContainsPoint(this.x, this.y+this.h) {
		return true
	}
	return false
}

func (this *Region) Clone() *Region {
	return &Region{this.x, this.y, this.w, this.h}
}

func (this *Region) String() string {
	return fmt.Sprintf("(%d, %d) -> (%d, %d)", this.x, this.y, this.w, this.h)
}

func (this *Region) Multiply(v int) {
	this.x *= v
	this.y *= v
	this.w *= v
	this.h *= v
}

type RegionMessage struct {
	MessageBase
	surface Surface
	regions []*Region
}

func newRegionMessage(messageType int, surface Surface, regions []*Region) *RegionMessage {
	return &RegionMessage{MessageBase{messageType}, surface, regions}
}

type VisibleAreaChangeMessage struct {
	MessageBase
	w, h int
}

func newVisibleAreaChangeMessage(w, h int) *VisibleAreaChangeMessage {
	return &VisibleAreaChangeMessage{MessageBase{MT_VisibleAreaChange}, w, h}
}

func _s_Fullscreen(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()
	ws.screen.setScreenMode(screenModeGraphic)

	return nil
}

func _s_Textscreen(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()
	ws.screen.setScreenMode(screenModeText)

	return nil
}

func _s_Splitscreen(frame Frame, parameters []Node) *CallResult {

	ws := frame.workspace()
	ws.screen.setScreenMode(screenModeSplit)

	return nil
}
