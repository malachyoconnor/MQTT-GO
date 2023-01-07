#!/bin/bash

for i in {1..1000}
do 
	mosquitto_pub -t "x/y/z" -m "$i !!" -p 8000
	
	if [ $(($i% 25)) == 0 ]; then
		mosquitto_sub -t "x/y/z" -p 8000 &
	fi

done
