#!/bin/bash
set -e

for file in ./include/mods/*/; do
	echo \[Mod Build\] Building $file
	if (cd "$file" && sh build.sh); then
	echo \[Mod Build\] Finished
	else
	echo \[Mod Build\] Failed to build $file
	fi
done

cp -r ./include/* ./_out/