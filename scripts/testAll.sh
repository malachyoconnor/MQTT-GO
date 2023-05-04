#!/bin/bash

go build .

for i in {20..200..20}
do
        ./MQTT-GO.exe test -clients $i -protocol $1

done

