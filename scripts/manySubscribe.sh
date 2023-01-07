#!/bin/bash

for i in {1..100}
do
	mosquitto_sub -t "x/$i" -p 8000 &
	echo "$i"
done
