#!/bin/bash

for i in {20..200..20}
do 

	./MQTT-GO.exe testLocalhost -clients $i -ip 45.156.85.191 -port 45457 -protocol UDP
	
done
