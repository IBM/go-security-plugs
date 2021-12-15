#!/bin/bash

echo "Build plugs:"
./buildPlugs.sh
echo "------------------------"
echo "Run Proxy with plugs"
go run .

