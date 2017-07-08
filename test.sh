#!/usr/bin/env bash

N=0
while true; do
   N=$(( $N + 1 ))
   echo "abc$N"
   sleep 5
done