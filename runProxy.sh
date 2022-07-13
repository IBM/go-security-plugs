#!/bin/bash

# Tel RTPLUGS which plugs to activate (comma seperated)
export RTPLUGS="rtgate,testgate"

# List of all supported plug packages (comma seperated)
RTPLUGS_PKG=""
RTPLUGS_PKG="${RTPLUGS_PKG},github.com/IBM/go-security-plugs/plugs/rtgate"
RTPLUGS_PKG="${RTPLUGS_PKG},github.com/IBM/go-security-plugs/plugs/testgate"


echo "------------------------"
echo "Generating auto_generate_imports.go"

cat <<EOT > auto_generate_imports.go
// Code generated by $0. DO NOT EDIT.

package main

EOT

IFS="," read -r -a PKG_ARRAY <<< ${RTPLUGS_PKG}
for p in ${PKG_ARRAY}
do
  # process
  echo import _ \"$p\"
  echo "import _ \"$p\"" >> auto_generate_imports.go
done

go mod tidy
go mod vendor

echo "------------------------"
echo "Run Proxy with static plugs"
go run .
