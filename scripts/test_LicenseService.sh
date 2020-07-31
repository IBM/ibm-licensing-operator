#!/bin/bash

echo "Tests the connection"
curl -k -s -w %{http_code} -X GET $1 -o /dev/null
