#!/bin/bash

clientNum="100"
protocol="TCP"

go build .

for i in {100..1000..100}

do 

	./MQTT-GO.exe test3 -clients 100 -ip 45.156.85.191 -port 44686 -protocol $protocol -packetSize $i
	
done
