#!/bin/bash

for i in {10..200..10}
do
	./MQTT-GO.exe test -clients $i -protocol UDP

done
