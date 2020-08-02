#!/bin/bash

echo "Tests the connection using the following URL:" $1
#export LS_URL="http://"$1$2
#echo "LS URL:"$LS_URL
curl -k -s -w %{http_code} -X GET $1 -o /dev/null
