#!/bin/bash

go build .

clientNum="100"
protocol="TCP"

./MQTT-GO test2 -ip 45.156.85.191 -port 44686 -protocol $protocol -packetSize 1 -packetNum 30000 -clients $clientNum
sleep 1

for i in {100..1000..100}
do 

	./MQTT-GO test2 -ip 45.156.85.191 -port 44686 -protocol $protocol -packetSize $i -packetNum 30000 -clients $clientNum
	sleep 1

done
