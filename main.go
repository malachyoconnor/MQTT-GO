package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/network"
	"MQTT-GO/stresstests"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	numFlags := permuteArgs()
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println(err)
			return
		}
		err = pprof.StartCPUProfile(f)
		if err != nil {
			fmt.Println("Error while starting profile", err)
			return
		}

		go func() {
			fmt.Println("STARTING PROFILING")
			time.Sleep(30 * time.Second)
			pprof.StopCPUProfile()
			fmt.Println("FINISHED PROFILING")
		}()
	}

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Please input gobro or client")
		return
	}

	switch args[numFlags] {
	case "gobro":
		{
			server := gobro.NewServer()
			server.StartServer()
		}
	case "client":
		{
			client.StartClient()
		}

	case "stresstest":
		{
			stresstests.ConnectAndPublish(200)
		}

	case "quic":
		{
			fmt.Println("Running QUIC test")
			network.RunTest()
		}

	default:
		{
			fmt.Println("Malformed input, exiting")
		}

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
