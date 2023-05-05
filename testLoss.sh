#!/bin/bash

go build .

for i in {5..30..5}
do 

	./MQTT-GO test1 -ip 45.156.85.191 -port 44686 -protocol UDP -packetSize 1 -packetNum 100000 -clients $i
	sleep 1

done
