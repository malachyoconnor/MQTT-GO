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
	"MQTT-GO/structures"
)

var (
	cpuprofile = flag.String("cpuprofile", "", "Profile code, and write that profile to a file")
	protocol   = flag.String("protocol", "TCP", "Select the transport protocol to use")
)

func main() {
	permuteArgs()
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			fmt.Println("Err while creating cpu profile:", err)
			fmt.Println("Attempting to save to:", *cpuprofile)
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
			for i := 0; i < 100; i++ {
				structures.PrintCentrally("FINISHED PROFILING")
			}
		}()
	}

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Please input gobro or client")
		return
	}

	connectionType, ok := map[string]byte{"TCP": 0, "QUIC": 1, "UDP": 2}[*protocol]
	if !ok {
		structures.Println("Malformed input, exiting")
		return
	}

	client.ConnectionType = connectionType
	gobro.ConnectionType = connectionType

	switch args[len(args)-1] {
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
			// stresstests.ManyClientsConnect(1000)
			stresstests.ManyClientsPublish(250)
		}

	case "quic":
		{
			fmt.Println("Serving")
			x := network.QUICListener{}
			x.Listen("127.0.0.1", 8000)
		}

	case "quicClient":
		{
			fmt.Println("Connecting")
			x := network.QUICCon{}
			go x.Connect("127.0.0.1", 8000)

			time.Sleep(500 * time.Millisecond)
			x.Write([]byte("Hello is this working?"))
			time.Sleep(1000 * time.Millisecond)
			x.Write([]byte("Hello is this working?"))
			time.Sleep(1000 * time.Millisecond)
		}

	default:
		{
			structures.Println("Malformed input, exiting")
		}

	}

}

// I want to be able to put non-options before the flags - to do this we permute the os.args
func permuteArgs() {

	for i := 1; i < len(os.Args)-1; i++ {
		os.Args[i], os.Args[i+1] = os.Args[i+1], os.Args[i]
	}

}
