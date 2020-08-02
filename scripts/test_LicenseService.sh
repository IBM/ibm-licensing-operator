#!/bin/bash

echo "Tests the connection using the following URL:" $1
curl -k -s -w %{http_code} -X GET $1 -o /dev/null
