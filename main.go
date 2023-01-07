package main

import (
	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {

	numFlags := permuteArgs()
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
		}
		pprof.StartCPUProfile(f)

		go func() {
			fmt.Println("STARTING PROFILING")
			time.Sleep(30 * time.Second)
			pprof.StopCPUProfile()
			fmt.Println("FINISHED PROFILING")
		}()
	}

	args := os.Args[1:]

	if args[numFlags] == "gobro" {
		server := gobro.CreateServer()
		server.StartServer()
	} else if args[numFlags] == "client" {
		client.StartClient()
	} else {
		fmt.Println("Malformed input, exiting")
	}

}

// I want to be able to put non-options before the flags - to do this we permute the os.args
func permuteArgs() (numFlags int) {

	args := os.Args[1:]

	for i := range args {
		if args[i][0] == '-' {
			args[i], args[numFlags] = args[numFlags], args[i]
			numFlags++
		}
	}
	return numFlags

}
