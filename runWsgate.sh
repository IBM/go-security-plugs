#!/bin/bash

echo "Build plugs:"
cd ./plugs/wsgate
go build -buildmode=plugin .
cd ../..
echo "------------------------"
echo "Run Proxy with plugs"
go run .

