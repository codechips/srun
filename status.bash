#!/bin/bash

duration=30
interval=1
steps=$((duration / interval))
width=50

for ((i=0; i<=steps; i++)); do
    sleep $interval
    progress=$((i * width / steps))
    printf "\r[%-${width}s] %d%%" $(printf "#%.0s" $(seq 1 $progress)) $((i * 100 / steps))
done

printf "\nDone!\n"
