package main

import (
	"github.com/adkennan/Go-SDL/sdl"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}

func runEventLoop(ws *Workspace) {

	s := ws.Screen()

	sdl.EnableUNICODE(1)

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN {
					ws.console.Input() <- &e.Keysym
				}
			}
		}
		s.Update()
		sdl.Delay(20)
	}
}

func init() {
}

func main() {

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		panic(sdl.GetError())
	}
	defer sdl.Quit()

	ws := CreateWorkspace()
	defer ws.Screen().Close()

	go ws.RunInterpreter()

	runEventLoop(ws)

}
