package gobro

import (
	"fmt"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var filter = ""
var iface = "lo"
var snaplen = 1024
var promisc = false
var timeoutT = 1

func Sniff(packetPool *BytePool) {

	fmt.Println("Sniffing")

	var timeout time.Duration = time.Duration(timeoutT) * time.Second
	handle, err := pcap.OpenLive(iface, int32(snaplen), promisc, timeout)
	if err != nil {
		fmt.Println(err)
	}

	defer handle.Close()

	// Applying BPF Filter if one exists
	if filter != "" {
		err := handle.SetBPFFilter(filter)
		if err != nil {
			fmt.Println("error applying BPF Filter: ", err)
		}
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// TODO: What we actually want is to check if the packet was meant for us.
	// 		 ATM we're just testing if it's an mqtt packet

	for packet := range packetSource.Packets() {

		// Check if this is a TCP packet
		if tcpLayer := packet.Layer(layers.LayerTypeTCP); tcpLayer == nil {
			continue
		}
		if packetPayload := packet.ApplicationLayer(); packetPayload == nil {
			continue
		}
		packetPool.Put(packet.ApplicationLayer().Payload())

	}

}
