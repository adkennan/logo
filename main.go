package main

import (
	"os"
	"strconv"
)

func main() {

	w := 0
	h := 0
	if len(os.Args) == 3 {
		var err error
		w, err = strconv.Atoi(os.Args[1])
		if err != nil {
			panic(err)
		}
		h, err = strconv.Atoi(os.Args[2])
		if err != nil {
			panic(err)
		}
	}

	ws := CreateWorkspace(w, h)

	ws.RunInterpreter()
}
