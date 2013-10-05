#!/bin/bash

dir="${0%/*}"
if [[ "${dir}" == "." || "${dir:0:1}" != "/" ]]; then
    dir=$(pwd)
fi

export GOPATH="${dir}:${GOPATH}"
