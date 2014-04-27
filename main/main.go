package main

import (
	"github.com/adkennan/Go-SDL/sdl"
	"github.com/adkennan/logo"
)

func runEventLoop(ws *logo.Workspace) {

	s := ws.Screen()
	for {
		sdl.Delay(20)

		s.Update()
	}
}

func main() {

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		panic(sdl.GetError())
	}
	defer sdl.Quit()

	ws := logo.CreateWorkspace()

	go runEventLoop(ws)

	ws.RunInterpreter()
}
