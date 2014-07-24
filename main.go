package main

import (
	"flag"
	"os"
	"runtime/pprof"
)

func main() {

	cpuprofile := flag.String("cpuprofile", "", "write profiling info to file.")
	w := flag.Int("w", 0, "screen width.")
	h := flag.Int("h", 0, "screen height.")

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	ws := CreateWorkspace()
	ws.OpenScreen(*w, *h)
	ws.RunInterpreter()
}
