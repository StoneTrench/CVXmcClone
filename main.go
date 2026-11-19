package main

import (
	"log/slog"

	"github.com/StoneTrench/CVXmcClone/src/client"
	"github.com/StoneTrench/CovarexEngine/engine/platform"
)

// //go:embed include/mods/*
// var FS_MODS embed.FS

func main() {
	var err error
	vfs := new(platform.VFS).Init(nil, true)
	// sub, err := fs.Sub(FS_MODS, "include")
	// if err != nil {
	// 	panic(err)
	// }
	// vfs.AppendFilesystem(sub)
	// fmt.Printf("vfs.Tree(): %v\n", vfs.TreeFS())

	err = client.Init(vfs)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
}
