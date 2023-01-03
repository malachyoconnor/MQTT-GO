package main

import (
	"MQTT-GO/gobro"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
		}
		pprof.StartCPUProfile(f)

		go func() {
			fmt.Println("STARTING PROFILING")
			time.Sleep(60 * time.Second)
			pprof.StopCPUProfile()
			fmt.Println("FINISHED PROFILING")
		}()
	}

	server := gobro.CreateServer()
	server.StartServer()

}
