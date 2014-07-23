package main

import (
	"flag"
	"os"
	"runtime/pprof"
	"strconv"
)

func main() {

	cpuprofile := flag.String("cpuprofile", "", "write profiling info to file.")

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

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	ws := CreateWorkspace(w, h)

	ws.RunInterpreter()
}
