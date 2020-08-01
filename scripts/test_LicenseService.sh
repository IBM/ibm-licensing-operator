#!/bin/bash

echo "Tests the connection"
export LS_URL="http://"$1$2
echo "LS URL:"$LS_URL
curl -k -s -w %{http_code} -X GET $LS_URL -o /dev/null
