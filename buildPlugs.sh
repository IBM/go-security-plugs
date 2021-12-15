#!/bin/bash

for dir in plugs/*/     # list directories in the form "/tmp/dirname/"
do
    echo "${dir}"    # print everything after the final "/"
    cd "${dir}"
    go build -buildmode=plugin .
    cd ../..
done
