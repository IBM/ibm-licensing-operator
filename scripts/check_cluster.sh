#!/bin/bash

EXPECTED_STATE='normal'
ACTUAL_STATE="$(ibmcloud ks cluster get --cluster $1 | grep ^State)"
ACTUAL_STATE="${ACTUAL_STATE#*:}"
echo "Expected state: " $EXPECTED_STATE
echo "Current state: " $ACTUAL_STATE

total_nb_of_minutes=0
time_to_wait=1m
until [ $ACTUAL_STATE == $EXPECTED_STATE ]
do
  echo "Waiting for " $time_to_wait " to check if cluster is in '" $EXPECTED_STATE "' state."
  sleep $time_to_wait
  ((total_nb_of_minutes++))
  ACTUAL_STATE="$(ibmcloud ks cluster get --cluster $1 | grep ^State)"
  ACTUAL_STATE="${ACTUAL_STATE#*:}"
  echo "... current state: " $ACTUAL_STATE
  EXPECTED_STATE="normal"
done

echo "Cluster turned into '" $EXPECTED_STATE "' state after " $total_nb_of_minutes " minute(s)"
