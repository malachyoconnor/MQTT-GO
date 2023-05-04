// Package main contains the main function for the project.
// It is used to start the broker, client, or stresstests.
// It contains flags to select the transport protocol, ip, and port.
// It also contains a flag to profile the code.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"runtime/pprof"
	"runtime/trace"

	_ "net/http/pprof"
	"os"
	"time"

	"MQTT-GO/client"
	"MQTT-GO/gobro"
	"MQTT-GO/stresstests"
	"MQTT-GO/structures"
)

var (
	cpuprofile  = flag.String("cpuprofile", "", "Profile code, and write that profile to a file")
	heapprofile = flag.String("heapprofile", "", "Profile code, and write that profile to a file")
	newTrace    = flag.String("trace", "", "Profile code, and write a trace fileto a file")
	protocol    = flag.String("protocol", "TCP", "Select the transport protocol to use")
	// PORT is the port to listen on for the server, or to connect to for the client
	PORT = flag.Int("port", 8000, "Select the port to use")
	// IP is the ip to listen on for the server, or to connect to for the client
	IP = flag.String("ip", "127.0.0.1", "Select the ip to use")
)

func main() {
	permuteArgs()
	flag.Parse()
	if *cpuprofile != "" {
		profileCPU()
	}
	if *heapprofile != "" {
		profileHeap()
	}
	if *newTrace != "" {
		createTrace()
	}

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("Please input gobro or client")
		return
	}

	connectionType, ok := map[string]byte{"TCP": 0, "QUIC": 1, "UDP": 2}[*protocol]
	if !ok {
		fmt.Println("Malformed input, exiting")
		return
	}

	fmt.Println("Protocol used:", *protocol)
	client.ConnectionType = connectionType
	gobro.ConnectionType = connectionType
	stresstests.ConnectionType = connectionType

	switch args[len(args)-1] {
	case "gobro":
		{
			server := gobro.NewServer()
			server.StartServer(*IP, *PORT)
		}
	case "client":
		{
			client.StartClient(*IP, *PORT)
		}
	case "stresstest":
		{
			// stresstests.ManyClientsConnect(*IP, *PORT)
			stresstests.ManyClientsPublish(*IP, *PORT, 1024*3)
		}
	case "test":
		{
			stresstests.TestManyClients()
		}
	default:
		{
			fmt.Println("Malformed input, exiting")
		}
	}
}

// I want to be able to put non-options before the flags - to do this we permute the os.args
func permuteArgs() {
	for i := 1; i < len(os.Args)-1; i++ {
		os.Args[i], os.Args[i+1] = os.Args[i+1], os.Args[i]
	}
}

func profileCPU() {
	file, err := os.Create(*cpuprofile)
	if err != nil {
		fmt.Println("Err while creating cpu profile:", err)
		fmt.Println("Attempting to save to:", *cpuprofile)
		return
	}

	err = pprof.StartCPUProfile(file)
	if err != nil {
		fmt.Println("Error while starting profile", err)
		return
	}
	go func() {
		fmt.Println("STARTING PROFILING")
		time.Sleep(15 * time.Second)
		pprof.StopCPUProfile()
		file.Close()
		for i := 0; i < 100; i++ {
			structures.PrintCentrally("FINISHED PROFILING")
		}
	}()
}

func profileHeap() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	file, err := os.Create(*heapprofile)
	fmt.Println("Profiling the heap")
	if err != nil {
		fmt.Println("Err while creating cpu profile:", err)
		fmt.Println("Attempting to save to:", *cpuprofile)
		return
	}
	go func() {
		defer file.Close()
		time.Sleep(50 * time.Second)
		for i := 0; i < 100; i++ {
			structures.PrintCentrally("FINISHED PROFILING")
		}
		err = pprof.WriteHeapProfile(file)
	}()
}

func createTrace() {
	fmt.Println("STARTING TRACE")
	f, err := os.Create("trace.out")
	if err != nil {
		panic(err)
	}

	go func() {
		defer f.Close()
		err = trace.Start(f)
		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 15)
		fmt.Println("FINISHING TRACE")
		defer trace.Stop()
	}()
}
