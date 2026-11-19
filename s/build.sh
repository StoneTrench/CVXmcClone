#!/bin/bash
set -e

rm -fr "./_out/"
mkdir -p "./_out/"
sh s/modbuild.sh

# go run github.com/StoneTrench/CovarexEngine/cli/build --os js --arch wasm
go run github.com/StoneTrench/CovarexEngine/cli/build