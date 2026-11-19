#!/bin/bash
set -e

go get github.com/StoneTrench/CovarexEngine/cli/gen/parser@v0.2.0-alpha
go run github.com/StoneTrench/CovarexEngine/cli/gen -l go -s bindings/env.abi -o bindings/